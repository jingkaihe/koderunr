var ROUTERS = {
  RUN: "/api/run/",
  SAVE: "/api/save/",
  STDIN: "/api/stdin/",
  REGISTER: "/api/register/",
  FETCH: "/api/fetch/"
}

$(function() {
  var term = $('#stdio').terminal(undefined, {
      name: 'KodeRunr',
      prompt: '> ',
      greetings: false,
  });

  var editor = ace.edit("editor");
  editor.setTheme("ace/theme/cobalt");
  editor.setOptions({
    fontSize: "10pt",
    tabSize: 2,
  });

  var KodeRunr = function(){
    this.term = term;
    this.term.focus(false);
    this.editor = ace.edit("editor");
    this.setLang($("#lang").val());
    this.running = false;
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
      case "JRuby":
        mode = "ruby";
        break;
      default:
        mode = this.lang;
    }
    this.editor.getSession().setMode("ace/mode/" + mode);
  };

  KodeRunr.prototype.runCode = function(evt) {
    // Do not run code when it's in the middle of running,
    // because it will make the console output messy (and
    // also confusing)
    if (this.running) {
      alert("The code is now running.\n\nYou can either refresh the page or wait for the finishing.")
      return
    }

    // Mark the runner as running.
    this.running = true;
    var sourceCode = this.editor.getValue();

    var runnable = { lang: this.lang, source: sourceCode };

    if (this.version) {
      runnable.version = this.version;
    }

    if (this.lang === "JRuby") {
      runnable.lang = "ruby";
      runnable.version = "jruby-" + this.version;
    }

    var runner = this;
    $.post(ROUTERS.REGISTER, runnable, function(uuid) {
      // Empty the output field
      runner.term.clear();
      runner.term.focus();
      var evtSource = new EventSource(ROUTERS.RUN + "?evt=true&uuid=" + uuid);
      evtSource.onmessage = function(e) {
        var str = e.data.split("\n").join("");
        runner.term.echo(str);
      }

      evtSource.onerror = function(e) {
        if (uuid) {
          uuid = undefined;
          runner.term.echo("[[;green;]Completed!]");
          runner.term.focus(false);
          runner.running = false;
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
      $("#lang").replaceWith("<span id='lang' class='lead'>" + lang + "</span>");
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
