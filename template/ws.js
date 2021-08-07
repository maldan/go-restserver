const RestWS = {
  ws: null,
  response: {},
  error: {},
  connect() {
    this.ws = new WebSocket(`ws://%HOST%/__ws`);
    this.ws.onmessage = (data) => {
      const msg = JSON.parse(data.data);
      if (msg.status) {
        this.response[msg.id](msg.response);
      } else {
        this.error[msg.id](msg.response);
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
};

RestWS.connect();
