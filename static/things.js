htmx.onLoad(function(content) {
  if (!(content.classList.contains("thing") && content.classList.contains("js"))) {
    return
  }

  let script = content.querySelector(".js-code");
  let output = content.querySelector(".js-output");
  let canvas = content.querySelector(".js-canvas");
  let ctx = canvas.getContext("2d");
  ctx.fillRect(Math.random()*canvas.width, Math.random()*canvas.height, 1, 1);

  let update = function() {
    try {
      output.classList.remove("js-error");
      output.textContent = JSON.stringify(eval("(function(canvas, ctx) {\n"+script.value+"\n})(canvas, ctx)"));
    } catch (err) {
      output.classList.add("js-error");
      output.textContent = err;
    }
  }

  if (script.value != "") {
    update();
  }

  script.addEventListener("keydown", function(ev) {
    if (ev.ctrlKey && ev.key == "Enter") {
      update();
    }
  })
  script.addEventListener("change", function() {
    update();
  });
});
