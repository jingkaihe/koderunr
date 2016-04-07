$(function() {
  var KodeRunr = function(ext){
    this.editor = ace.edit("editor");
    this.editor.setTheme("ace/theme/monokai");
    this.editor.setOptions({
      fontSize: "12pt",
    });

    this.setExt(ext);
  }

  KodeRunr.prototype.LANG_MAPPING = {
    ".go": "golang",
    ".rb": "ruby",
    ".c": "c_cpp",
    ".ex": "elixir",
  };

  KodeRunr.prototype.setExt = function(ext) {
    this.ext = ext;
    this.editor.getSession().setMode("ace/mode/" + this.LANG_MAPPING[this.ext]);
  };

  KodeRunr.prototype.runCode = function(evt) {
    var sourceCode = this.editor.getValue();

    var runnable = { ext: this.ext, source: sourceCode };
    if (this.version) {
      runnable.version = this.version
    }
    $.post('/register/', runnable, function(uuid) {
      // Empty the output field
      $("#streamingResult").text("");
      $("#inputField").val("").focus();

      var evtSource = new EventSource("/run?evt=true&uuid=" + uuid);
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
          $.post('/stdin/', {
            input: input,
            uuid: uuid
          });
        }
      });
    });
  };

  var sourceCodeCache = sourceCodeCache || {};
  sourceCodeCache.fetch = function(runner) {
    return localStorage.getItem(runner.ext)
  }

  sourceCodeCache.store = function(runner){
    localStorage.setItem(runner.ext, runner.editor.getValue())
  }

  var runner = new KodeRunr($("#ext").val());

  $("#submitCode").on("click", runner.runCode.bind(runner));

  $(document).on("keydown", function(e){
    if (e.keyCode == 13 && (e.ctrlKey || e.metaKey)) {
       runner.runCode();
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
