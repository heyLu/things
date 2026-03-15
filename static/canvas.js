class SVGPath2D extends Path2D{
  constructor(parts = []) {
    super(parts.join(' '));

    this.parts = parts;
  }

  lineTo(x, y) {
    if (this.parts.length == 0) {
      this.parts.push(`M${x} ${y}`);
    }
    this.parts.push(`L${x} ${y}`);

    super.lineTo(x, y);
  }

  toSVGString() {
    return this.parts.join(' ');
  }
}

class Canvas {
  constructor(canvas, window = null) {
    this.canvas = canvas;

    this.drawMode = document.createElement("input");
    this.drawMode.type = "checkbox";
    this.drawMode.style = "position: fixed; top: 0; right: 0;";
    this.canvas.parentElement.appendChild(this.drawMode);

    if (window) {
      this.canvas.width = window.innerWidth;
      this.canvas.height = window.innerHeight;
    } else {
      let rect = canvas.getBoundingClientRect();
      this.canvas.width = rect.width;
      this.canvas.height = rect.height;
    }
    this.context = canvas.getContext('2d');

    this.action = null;
    this.lastEv = null;
    this.path = null;

    this.worldPos = {x: 0, y: 0};
    this.pixelPos = {
      offsetX: this.canvas.width / 2,
      offsetY: this.canvas.height / 2,
    };

    this.objects = [];
    if (localStorage.hasOwnProperty("objects")) {
      let oldObjects = JSON.parse(localStorage.getItem("objects"));
      for (let obj of oldObjects) {
        switch (obj.type) {
          case "path":
            this.objects.push({type: "path", path: new SVGPath2D(obj.path.parts)});
            break;
          default:
            this.objects.push(obj);
        }
      }
    }

    this.setupEvents();
  }

  setupEvents() {
    let self = this;

    this.canvas.addEventListener("pointermove", (ev) => {
      if (self.action == "draw") {
        let pos = this.pixelToWorld(ev.offsetX, ev.offsetY);
        self.path.lineTo(pos.x, pos.y);
      }

      self.draw(ev);
    });

    this.canvas.addEventListener("pointerdown", (ev) => {
      let action = null;
      switch (ev.pointerType) {
        case "mouse":
          if (ev.button == 1) {
            action = "move";
          } else if (ev.button == 0) {
            action = "draw";
            self.path = new SVGPath2D();
          }
          break;
      }

      if (action == null) {
        return;
      }

      self.action = action;
      self.lastEv = ev;
    });
    this.canvas.addEventListener("touchstart", (ev) => {
      let action = null;
      if (ev.touches.length == 1 && !self.drawMode.checked) {
        action = "draw";
        self.path = new SVGPath2D();
      } else if (ev.touches.length == 2 || self.drawMode.checked) {
        action = "move";
        // FIXME: get diff from touches ...
      }

      if (action == null) {
        return;
      }

      self.action = action;
      self.lastEv = ev;
    });

    this.canvas.addEventListener("pointerup", (ev) => {
      switch (self.action) {
        case "move":
          self.moveBy(ev.offsetX, ev.offsetY);
          break;
        case "draw":
          self.objects.push({type: "path", path: self.path});

          localStorage.setItem("objects", JSON.stringify(self.objects));

          self.path = null;
          self.action = null;
          self.lastEv = null;
        case null:
          break;
        default:
          console.error(`unknown action ${self.action}`);
      }
    });

    this.canvas.addEventListener("pointerleave", (ev) => {
      switch (self.action) {
        case "move":
          self.moveBy(ev.offsetX, ev.offsetY);
          break;
        default:
          console.error(`unknown action ${self.action}`);
      }

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

    this.action = null;
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
    if (this.lastEv && this.action == "move") {
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

    for (let obj of this.objects) {
      switch (obj.type) {
        case "rect":
          this.context.fillRect(obj.x, obj.y, obj.width, obj.height);
          break;
        case "path":
          this.context.stroke(obj.path);
          break;
        default:
          console.error("unknown object type", obj.type);
      }
    }

    if (this.action == "draw") {
      this.context.stroke(this.path);
    }

    this.context.restore();

    let world = this.pixelToWorld(ev.offsetX, ev.offsetY);
    let text = `${world.x},${world.y}`;
    let textSize = this.context.measureText(text);
    this.context.fillText(text, this.canvas.width - textSize.width, this.canvas.height - textSize.actualBoundingBoxAscent);    
  }
}
