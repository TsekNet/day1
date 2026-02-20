(function() {
  "use strict";

  var Backend = null;
  var allPages = [];
  var currentIndex = 0;
  var totalPages = 0;
  var onFinalPage = false;

  function findApp() {
    if (!window.go) return null;
    for (var ns in window.go) {
      if (window.go[ns] && window.go[ns].App) return window.go[ns].App;
    }
    return null;
  }

  function applyTheme(theme) {
    if (theme === "dark") {
      document.documentElement.setAttribute("data-theme", "dark");
    } else if (theme === "light") {
      document.documentElement.setAttribute("data-theme", "light");
    } else {
      // "auto" — check matchMedia, fall through to CSS media query if supported
      if (window.matchMedia && window.matchMedia("(prefers-color-scheme: dark)").matches) {
        document.documentElement.setAttribute("data-theme", "dark");
      }
    }
  }

  function init() {
    Backend = findApp();
    if (!Backend) return;

    Backend.GetTheme().then(applyTheme);

    Backend.GetAccentColor().then(function(color) {
      if (color && /^#([0-9A-Fa-f]{3}|[0-9A-Fa-f]{6})$/.test(color)) {
        if (color.length === 4) {
          color = "#" + color[1]+color[1] + color[2]+color[2] + color[3]+color[3];
        }
        var r = document.documentElement;
        r.style.setProperty("--accent", color);
        r.style.setProperty("--accent-dim", color);
        r.style.setProperty("--accent-soft", color + "1a");
      }
    });

    Backend.GetBrand().then(function(brand) {
      if (brand && brand.name) {
        var container = document.getElementById("brand");
        container.style.display = "flex";
        if (brand.logo) {
          var img = document.createElement("img");
          img.className = "brand-logo";
          img.src = brand.logo;
          img.alt = brand.name;
          container.appendChild(img);
        }
        var name = document.createElement("span");
        name.className = "brand-name";
        name.textContent = brand.name;
        container.appendChild(name);
      }
    });

    Backend.GetHelpURL().then(function(url) {
      if (url) {
        var meta = document.querySelector(".footer-meta");
        var link = document.getElementById("help-link");
        var divider = document.createElement("div");
        divider.className = "footer-divider";
        meta.insertBefore(divider, link);
        link.style.display = "inline";
        link.addEventListener("click", function(e) {
          e.preventDefault();
          Backend.OpenHelp();
        });
      }
    });

    Backend.GetPages().then(function(pages) {
      allPages = pages || [];
      totalPages = allPages.length;
      buildProgress();
      showPage(0);
      Backend.Ready();
    });
  }

  function buildProgress() {
    var container = document.getElementById("progress");
    container.innerHTML = "";

    for (var i = 0; i < totalPages; i++) {
      if (i > 0) {
        var line = document.createElement("div");
        line.className = "step-line";
        line.setAttribute("data-line-index", i - 1);
        container.appendChild(line);
      }

      var step = document.createElement("div");
      step.className = "step";
      step.setAttribute("data-step-index", i);
      step.addEventListener("click", onStepClick);

      var dot = document.createElement("div");
      dot.className = "step-dot";
      step.appendChild(dot);

      var label = document.createElement("span");
      label.className = "step-label";
      label.textContent = allPages[i].title || "Step " + (i + 1);
      step.appendChild(label);

      container.appendChild(step);
    }
  }

  function onStepClick(e) {
    var step = e.currentTarget;
    var index = parseInt(step.getAttribute("data-step-index"), 10);
    if (isNaN(index) || index < 0 || index >= totalPages) return;
    showPage(index);
  }

  function updateProgress() {
    var steps = document.querySelectorAll(".step");
    var lines = document.querySelectorAll(".step-line");

    for (var i = 0; i < steps.length; i++) {
      steps[i].classList.remove("active", "completed");
      if (onFinalPage) {
        steps[i].classList.add("completed");
      } else if (i < currentIndex) {
        steps[i].classList.add("completed");
      } else if (i === currentIndex) {
        steps[i].classList.add("active");
      }
    }

    for (var j = 0; j < lines.length; j++) {
      lines[j].classList.remove("completed");
      if (onFinalPage || j < currentIndex) {
        lines[j].classList.add("completed");
      }
    }
  }

  function showPage(index) {
    currentIndex = index;
    onFinalPage = false;

    Backend.GetPageHTML(index).then(function(html) {
      var content = document.getElementById("content");
      content.className = "content";
      content.innerHTML = html;

      var indicator = document.getElementById("page-indicator");
      indicator.textContent = (index + 1) + " of " + totalPages;

      var btnNext = document.getElementById("btn-next");
      btnNext.textContent = (index === totalPages - 1) ? "Finish" : "Next";
      document.getElementById("btn-close").style.display = "";

      updateProgress();
    });
  }

  function showFinalPage() {
    onFinalPage = true;
    updateProgress();

    Backend.GetFinalHTML().then(function(html) {
      var content = document.getElementById("content");
      if (html) {
        content.className = "content";
        content.innerHTML = html;
      } else {
        content.className = "content final-page";
        content.innerHTML =
          '<div class="final-check">' +
            '<svg viewBox="0 0 24 24"><polyline points="20 6 9 17 4 12"></polyline></svg>' +
          '</div>' +
          '<h1>You\'re all set!</h1>' +
          '<p>You\'re ready to go. Close this window to get started.</p>';
      }

      document.getElementById("page-indicator").textContent = "";
      document.getElementById("btn-close").style.display = "none";
      document.getElementById("btn-next").textContent = "Close";
    });
  }

  function advance() {
    if (onFinalPage) {
      Backend.Complete();
      return;
    }
    if (currentIndex < totalPages - 1) {
      showPage(currentIndex + 1);
    } else {
      showFinalPage();
    }
  }

  document.getElementById("btn-next").addEventListener("click", advance);

  document.getElementById("btn-close").addEventListener("click", function() {
    Backend.Dismiss();
  });

  document.addEventListener("keydown", function(e) {
    if (e.key === "Enter") {
      e.preventDefault();
      advance();
    } else if (e.key === "Backspace") {
      e.preventDefault();
      if (!onFinalPage && currentIndex > 0) {
        showPage(currentIndex - 1);
      }
    } else if (e.key === "Escape") {
      Backend.Dismiss();
    }
  });

  if (document.readyState === "loading") {
    document.addEventListener("DOMContentLoaded", init);
  } else {
    init();
  }
})();
