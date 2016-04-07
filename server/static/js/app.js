$(function() {
  var KodeRunr = function(ext){
    this.editor = ace.edit("editor");
    this.editor.setTheme("ace/theme/monokai");
    this.editor.setOptions({
      fontSize: "12pt",
    });

    this.setExt(ext);
  }

  KodeRunr.LANG_MAPPING = {
    ".go": "golang",
    ".rb": "ruby",
    ".c": "c_cpp",
    ".ex": "elixir",
  };

  KodeRunr.ROUTES = {
    RUN: "/run/",
    SAVE: "/save/",
    STDIN: "/stdin/",
    REGISTER: "/register/",
  }

  KodeRunr.prototype.setExt = function(ext) {
    this.ext = ext;
    this.editor.getSession().setMode("ace/mode/" + KodeRunr.LANG_MAPPING[this.ext]);
  };

  KodeRunr.prototype.runCode = function(evt) {
    var sourceCode = this.editor.getValue();

    var runnable = { ext: this.ext, source: sourceCode };

    if (this.version) {
      runnable.version = this.version;
    }

    $.post(KodeRunr.ROUTES.REGISTER, runnable, function(uuid) {
      // Empty the output field
      $("#streamingResult").text("");
      $("#inputField").val("").focus();

      var evtSource = new EventSource(KodeRunr.ROUTES.RUN + "?evt=true&uuid=" + uuid);
      evtSource.onmessage = function(e) {
        var text = $("#streamingResult").text();
        $("#streamingResult").text(text + e.data);
      }

      $("#inputField").on("keydown", function(evt){
        // Disable the arrow keys
        if([37, 38, 39, 40].indexOf(evt.which) > -1) {
            evt.preventDefault();
        }

        if (evt.which == 13) {
          var text = $(this).val();
          var lastCarriageReturn = text.lastIndexOf("\n")
          var input;
          if (lastCarriageReturn == -1) {
            input = text + "\n"
          }else{
            input = text.substr(lastCarriageReturn, text.length) + "\n"
          }
          $.post(KodeRunr.ROUTES.STDIN, {
            input: input,
            uuid: uuid
          });
        }
      });
    });
  };

  KodeRunr.prototype.saveCode = function(event) {
    var sourceCode = this.editor.getValue();

    var runnable = { ext: this.ext, source: sourceCode };
    if (this.version) {
      runnable.version = this.version
    }

    $.post(KodeRunr.ROUTES.SAVE, runnable, function(uuid) {
      alert(uuid);
    });
  }

  var sourceCodeCache = sourceCodeCache || {};
  sourceCodeCache.fetch = function(runner) {
    return localStorage.getItem(runner.ext)
  }

  sourceCodeCache.store = function(runner){
    localStorage.setItem(runner.ext, runner.editor.getValue())
  }

  var runner = new KodeRunr($("#ext").val());

  $("#submitCode").on("click", runner.runCode.bind(runner));
  $("#shareCode").on("click", runner.saveCode.bind(runner));

  // Shortcuts
  $(document).on("keydown", function(e){
    if (e.ctrlKey || e.metaKey) {
      switch (e.keyCode) {
      // run
      case 13:
        runner.runCode();
        break;
      // save
      case 83:
        e.preventDefault()
        runner.saveCode();
        break;
      }
    }
  });

  $("#ext").on("change", function() {
    // Empty the screen
    sourceCodeCache.store(runner)
    runner.editor.setValue("", undefined);
    $("#streamingResult").text("");

    var [ext, version] = this.value.split(" ")
    runner.setExt(ext);

    runner.version = version

    var cachedSourceCode = sourceCodeCache.fetch(runner)
    if (cachedSourceCode) {
      runner.editor.setValue(cachedSourceCode, 1);
    }
  });
});
