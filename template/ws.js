const __writeString = (dataView, offset, str) => {
  for (let i = 0; i < str.length; i++) {
    dataView.setUint8(offset + i, str.charCodeAt(i));
  }
};
const __writeArray = (dataView, offset, arr) => {
  for (let i = 0; i < arr.length; i++) {
    dataView.setUint8(offset + i, arr[i]);
  }
};
const __readString = (dataView, offset, len) => {
  let out = "";
  for (let i = 0; i < len; i++) {
    out += String.fromCharCode(dataView.getUint8(offset + i));
  }
  return out;
};

const RestWS = {
  ws: null,
  response: {},
  error: {},
  eventList: {},
  connect() {
    this.ws = new WebSocket(`ws://%HOST%/__ws`);
    this.ws.onmessage = async (data) => {
      const msg = {};
      if (typeof data.data === "string") {
        msg = JSON.parse(data.data);
      } else {
        var buff = await data.data.arrayBuffer();
        var dataView = new DataView(buff);
        msg.id = __readString(dataView, 0, 8);
        msg.status = dataView.getUint8(8, 1) === 1;
        if (!msg.status) {
          msg.response = __readString(dataView, 9, buff.byteLength - 9);
        } else {
          msg.response = new Uint8Array(buff).slice(9);
        }
      }

      if (msg.event) {
        if (this.eventList[msg.event]) {
          for (let i = 0; i < this.eventList[msg.event].length; i++) {
            this.eventList[msg.event][i](msg.response);
          }
        }
      } else {
        if (msg.status) {
          this.response[msg.id](msg.response);
        } else {
          this.error[msg.id](msg.response);
        }
      }
    };
    this.ws.onclose = () => {
      setTimeout(() => {
        this.connect();
      }, 1000);
    };
  },
  send(method, args) {
    const id = Math.random() + "";

    this.ws.send(
      JSON.stringify({
        id: id,
        method,
        args,
      })
    );

    return new Promise((resolve, reject) => {
      this.response[id] = resolve;
      this.error[id] = reject;
    });
  },
  sendBin(method, data) {
    let id = (Math.random() + "").slice(0, 8);
    var buffer = new ArrayBuffer(8 + 2 + method.length + 4 + data.length);
    var dataView = new DataView(buffer);

    // Set package id length
    __writeString(dataView, 0, id);

    // Set method name length
    dataView.setUint16(id.length, method.length);
    __writeString(dataView, id.length + 2, method);

    // Set data length
    dataView.setUint32(id.length + 2 + method.length, data.length);
    __writeArray(dataView, id.length + 2 + method.length + 4, data);

    // Send to server
    this.ws.send(new Uint8Array(buffer));

    return new Promise((resolve, reject) => {
      this.response[id] = resolve;
      this.error[id] = reject;
    });
  },
  on(event, fn) {
    if (!this.eventList[event]) {
      this.eventList[event] = [];
    }
    this.eventList[event].push(fn);
  },
};

RestWS.connect();
