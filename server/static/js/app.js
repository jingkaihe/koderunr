var ROUTERS = {
  RUN: "/run/",
  SAVE: "/save/",
  STDIN: "/stdin/",
  REGISTER: "/register/",
  FETCH: "/fetch/"
}

$(function() {
  var editor = ace.edit("editor");
  editor.setTheme("ace/theme/monokai");
  editor.setOptions({
    fontSize: "12pt",
  });

  var KodeRunr = function(){
    this.editor = ace.edit("editor");
    this.setLang($("#lang").val());
  }

  KodeRunr.prototype.setLang = function(lang) {
    langs = lang.split(" ");
    this.lang = langs[0];
    this.version = langs[1];

    var mode
    switch (this.lang) {
      case "go":
        mode = "golang";
        break;
      case "c":
        mode = "c_cpp";
        break;
      default:
        mode = this.lang;
    }
    this.editor.getSession().setMode("ace/mode/" + mode);
  };

  KodeRunr.prototype.runCode = function(evt) {
    var sourceCode = this.editor.getValue();
    var runnable = { lang: this.lang, source: sourceCode };

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

    var runnable = { lang: this.lang, source: sourceCode };
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
    return localStorage.getItem(runner.lang)
  }

  sourceCodeCache.store = function(runner){
    localStorage.setItem(runner.lang, runner.editor.getValue())
  }

  var runner = new KodeRunr();
  var codeID = window.location.hash.substring(1);

  if (codeID) {
    $.get(ROUTERS.FETCH + "?codeID=" + codeID, function(data) {
      var lang = data.lang;
      if (data.version) {
        lang = lang + " " + data.version;
      }

      $("#lang").val(lang);
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

  $("#lang").on("change", function() {
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
