class Canvas {
  constructor(canvas, window = null) {
    this.canvas = canvas;

    if (window) {
      this.canvas.width = window.innerWidth;
      this.canvas.height = window.innerHeight;
    } else {
      let rect = canvas.getBoundingClientRect();
      this.canvas.width = rect.width;
      this.canvas.height = rect.height;
    }

    this.context = canvas.getContext('2d');
    this.lastEv = null;

    this.worldPos = {x: 0, y: 0};
    this.pixelPos = {
      offsetX: this.canvas.width / 2,
      offsetY: this.canvas.height / 2,
    };

    this.setupEvents();
  }

  setupEvents() {
    let self = this;

    this.canvas.addEventListener("pointermove", (ev) => { self.draw(ev) });

    this.canvas.addEventListener("pointerdown", (ev) => {
      if (ev.pointerType == "mouse" && ev.button != 1) {
        return;
      }

      self.lastEv = ev;
    });

    this.canvas.addEventListener("pointerup", (ev) => {
      self.moveBy(ev.offsetX, ev.offsetY);
    });

    this.canvas.addEventListener("pointerleave", (ev) => {
      self.moveBy(ev.offsetX, ev.offsetY);

      self.draw(ev);
    });
  }

  moveBy(x, y) {
    if (!this.lastEv) {
      return;
    }

    this.pixelPos.offsetX -= this.lastEv.offsetX - x;
    this.pixelPos.offsetY -= this.lastEv.offsetY - y;

    this.worldPos.x -= this.lastEv.offsetX - x;
    this.worldPos.y -= this.lastEv.offsetY - y;

    this.lastEv = null;
  }

  pixelToWorld(x, y) {
    return {
      x: x - this.pixelPos.offsetX,
      y: y - this.pixelPos.offsetY,
    };
  }

  draw(ev = {offsetX: 0, offsetY: 0}) {
    this.context.clearRect(0, 0, this.canvas.width, this.canvas.height);

    let offsetX = 0;
    let offsetY = 0;
    if (this.lastEv) {
      offsetX = this.lastEv.offsetX - ev.offsetX;
      offsetY = this.lastEv.offsetY - ev.offsetY;
    }

    this.context.save();
    this.context.translate(this.pixelPos.offsetX - offsetX, this.pixelPos.offsetY - offsetY);

    this.context.save();
    this.context.fillStyle = "#777";
    this.context.beginPath();
    const dotSize = 1.3;
    const gridSize = 200;
    const maxWidth = Math.round(this.canvas.width / 2 / gridSize) * gridSize;
    const maxHeight = Math.round(this.canvas.height / 2 / gridSize) * gridSize;
    for (let x = -maxWidth; x <= maxWidth; x += gridSize) {
      for (let y = -maxHeight; y <= maxHeight; y += gridSize) {
        this.context.moveTo(x, y);
        let size = dotSize;
        if (x % 1000 == 0 && y % 1000 == 0) {
          size = 2;
        }
        this.context.ellipse(x, y, size, size, 0, 0, 360);   
      }
    }
    this.context.closePath();
    this.context.fill();
    this.context.restore();

    // for (let obj of objects) {
    //   switch (obj.type) {
    //     case "rect":
    //       context.fillRect(obj.x, obj.y, obj.width, obj.height);
    //       break;
    //     default:
    //       console.error("unknown object type", obj.type);
    //   }
    // }

    this.context.restore();

    let world = this.pixelToWorld(ev.offsetX, ev.offsetY);
    let text = `${world.x},${world.y}`;
    let textSize = this.context.measureText(text);
    this.context.fillText(text, this.canvas.width - textSize.width, this.canvas.height - textSize.actualBoundingBoxAscent);    
  }
}
