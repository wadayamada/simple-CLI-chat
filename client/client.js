const readline = require("readline");
const io = require("socket.io-client");
const socket = io("http://localhost:5000");

const user = process.argv[2];

rl = readline.createInterface(process.stdin, process.stdout);
rl.setPrompt("");

rl.on("line", (line) => {
  socket.emit("chat message", user, line);
}).on("close", () => {
  process.exit(0);
});

socket.on("connect", () => {
  console.log(`connected, id: ${socket.id}`);
  socket.emit("archive", user);
});

socket.on("archive", (archives) => {
  console.log(`archives: ${archives}`);
});

socket.on("broadcast", (data) => {
  console.log(data);
});
