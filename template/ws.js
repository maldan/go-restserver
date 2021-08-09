const RestWS = {
  ws: null,
  response: {},
  error: {},
  eventList: {},
  connect() {
    this.ws = new WebSocket(`ws://%HOST%/__ws`);
    this.ws.onmessage = (data) => {
      const msg = JSON.parse(data.data);
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
  on(event, fn) {
    if (!this.eventList[event]) {
      this.eventList[event] = [];
    }
    this.eventList[event].push(fn);
  },
};

RestWS.connect();
