let canvas = document.getElementById("canvas");

let rect = canvas.getBoundingClientRect();
canvas.width = rect.width;
canvas.height = rect.height;

let context = canvas.getContext("2d");

let worldPos = {x: 0, y: 0};
let pixelPos = {
  offsetX: canvas.width / 2,
  offsetY: canvas.height / 2,
  zoom: 1,
};
let objects = [{type: "rect", x: -100, y: -50, width: 10, height: 10}];
let lastEv = null;

draw({offsetX: 0, offsetY: 0});

function draw(ev) {
  context.clearRect(0, 0, canvas.width, canvas.height);

  let offsetX = 0;
  let offsetY = 0;
  if (lastEv) {
    offsetX = lastEv.offsetX - ev.offsetX;
    offsetY = lastEv.offsetY - ev.offsetY;
  }

  context.save();
  context.translate(pixelPos.offsetX + offsetX, pixelPos.offsetY + offsetY);
  context.fillRect(0, 0, 1, 1);

  for (let obj of objects) {
    switch (obj.type) {
      case "rect":
        context.fillRect(obj.x, obj.y, obj.width, obj.height);
        break;
      default:
        console.error("unknown object type", obj.type);
    }
  }

  context.restore();

  let text = `${worldPos.x},${worldPos.y} ${ev.offsetX},${ev.offsetY}`;
  let textSize = context.measureText(text);
  context.fillText(text, canvas.width - textSize.width, canvas.height - textSize.actualBoundingBoxAscent);
}

canvas.addEventListener("pointermove", (ev) => {
  draw(ev);
});

canvas.addEventListener("pointerdown", (ev) => {
  if (ev.pointerType == "mouse" && ev.button != 0) {
    return;
  }

  lastEv = ev;
});

canvas.addEventListener("pointerup", (ev) => {
  if (!lastEv) {
    return;
  }

  pixelPos.offsetX += lastEv.offsetX - ev.offsetX;
  pixelPos.offsetY += lastEv.offsetY - ev.offsetY;

  lastEv = null;
});

canvas.addEventListener("pointerleave", (ev) => {
  if (!lastEv) {
    return;
  }

  pixelPos.offsetX += lastEv.offsetX - ev.offsetX;
  pixelPos.offsetY += lastEv.offsetY - ev.offsetY;

  lastEv = null;

  draw(ev);
});
