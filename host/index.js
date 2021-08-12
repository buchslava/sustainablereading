const express = require("express");
const app = express();
const port = 3100;

const QTY_LIMIT = 3;
const TIME_LIMIT = 30;
const times = [];
const seconds = (d) => d.getTime() / 1000;

function allow() {
  const now = seconds(new Date());
  let qty = 0;

  for (let i = times.length - 1; i >= 0; i--) {
    if (now - seconds(times[i]) < TIME_LIMIT) {
      qty++;
    }
  }

  return qty <= QTY_LIMIT;
}

for (let i = 1; i < 100; i++) {
  app.get(`/data${i}`, (req, res) => {
    if (!allow()) {
      return res.status(500).send({ error: "denied!" });
    }
    times.push(new Date());
    res.send(`${Math.random()}`);
  });
}

app.listen(port, () => {
  console.log(`http://localhost:${port}`);
});
