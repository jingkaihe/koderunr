var ROUTERS = {
  RUN: "/run/",
  SAVE: "/save/",
  STDIN: "/stdin/",
  REGISTER: "/register/",
  FETCH: "/fetch/"
}

var LANG_MAPPING = {
  ".go": "golang",
  ".rb": "ruby",
  ".c": "c_cpp",
  ".ex": "elixir",
};

$(function() {
  var KodeRunr = function(){
    this.defaultEditor();
    this.setLang($("#ext").val());
  }

  KodeRunr.prototype.defaultEditor = function() {
    this.editor = ace.edit("editor");
    this.editor.setTheme("ace/theme/monokai");
    this.editor.setOptions({
      fontSize: "12pt",
    });
  };

  KodeRunr.prototype.setLang = function(lang) {
    [this.ext, this.version] = lang.split(" ")
    this.editor.getSession().setMode("ace/mode/" + LANG_MAPPING[this.ext]);
  };

  KodeRunr.prototype.runCode = function(evt) {
    var sourceCode = this.editor.getValue();
    var runnable = { ext: this.ext, source: sourceCode };

    if (this.version) {
      runnable.version = this.version;
    }

    $.post(ROUTERS.REGISTER, runnable, function(uuid) {
      // Empty the output field
      $("#streamingResult").text("");
      $("#inputField").val("").focus();

      var evtSource = new EventSource(ROUTERS.RUN + "?evt=true&uuid=" + uuid);
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
          $.post(ROUTERS.STDIN, {
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

    if (this.codeID) {
      runnable.codeID = this.codeID;
    }

    $.post(ROUTERS.SAVE, runnable, function(codeID) {
      window.history.pushState(codeID, "KodeRunr#" + codeID, "/#" + codeID);
    });
  }

  var sourceCodeCache = sourceCodeCache || {};
  sourceCodeCache.fetch = function(runner) {
    return localStorage.getItem(runner.ext)
  }

  sourceCodeCache.store = function(runner){
    localStorage.setItem(runner.ext, runner.editor.getValue())
  }

  var runner = new KodeRunr();
  var codeID = window.location.hash.substring(1);

  if (codeID) {
    $.get(ROUTERS.FETCH + "?codeID=" + codeID, function(msg) {
      var data = JSON.parse(msg);
      var lang = data.ext;
      if (data.version) {
        lang = lang + " " + data.version;
      }
      $("#ext").val(lang);
      runner.setLang(lang);
      runner.editor.setValue(data.source, 1);
      runner.codeID = codeID;
    });
  }

  $("#submitCode").on("click", function(event){
    runner.runCode();
  });

  $("#shareCode").on("click", function(event){
    runner.saveCode();
  });

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
        e.preventDefault();
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

    runner.setLang(this.value);

    var cachedSourceCode = sourceCodeCache.fetch(runner);

    if (cachedSourceCode) {
      runner.editor.setValue(cachedSourceCode, 1);
    }
  });
});
