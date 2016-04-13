var ROUTERS = {
  RUN: "/run/",
  SAVE: "/save/",
  STDIN: "/stdin/",
  REGISTER: "/register/",
  FETCH: "/fetch/"
}

$(function() {
  var term = $('#stdio').terminal(undefined, {
      name: 'KodeRunr',
      height: 200,
      prompt: '> ',
      greetings: false,
  });

  var editor = ace.edit("editor");
  editor.setTheme("ace/theme/monokai");
  editor.setOptions({
    fontSize: "12pt",
  });

  var KodeRunr = function(){
    this.term = term;
    this.term.focus(false);
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

    var runner = this;
    $.post(ROUTERS.REGISTER, runnable, function(uuid) {
      // Empty the output field
      runner.term.clear();
      runner.term.focus();
      var evtSource = new EventSource(ROUTERS.RUN + "?evt=true&uuid=" + uuid);
      evtSource.onmessage = function(e) {
        var data = e.data;
        var str = data.substring(0, data.length - 1);
        runner.term.echo(str);
      }

      evtSource.onerror = function(e) {
        if (uuid) {
          uuid = undefined;
          runner.term.echo("[[;green;]Completed!]");
          runner.term.focus(false);
        }
      }
      // Get the command and send to stdin.
      runner.term.on("keydown", function(e){
        if (uuid) {
          if (e.keyCode == 13) {
            var cmd = runner.term.get_command() + "\n";
            $.post(ROUTERS.STDIN, {
              input: cmd,
              uuid: uuid,
            });
          }
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
    runner.term.clear();

    runner.setLang(this.value);

    var cachedSourceCode = sourceCodeCache.fetch(runner);

    if (cachedSourceCode) {
      runner.editor.setValue(cachedSourceCode, 1);
    }
  });
});
