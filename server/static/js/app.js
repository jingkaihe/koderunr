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
  };

  KodeRunr.prototype.setExt = function(ext) {
    this.ext = ext;
    this.editor.getSession().setMode("ace/mode/" + this.LANG_MAPPING[this.ext]);
  };

  KodeRunr.prototype.runCode = function(evt) {
    var sourceCode = this.editor.getValue();

    $.post('/register/', { ext: this.ext, source: sourceCode }, function(msg) {
      // Empty the output field
      $("#streamingResult").text("");
      var evtSource = new EventSource("/run?evt=true&uuid=" + msg);
      evtSource.onmessage = function(e) {
        var text = $("#streamingResult").text();
        $("#streamingResult").text(text + e.data);
      }
    });
  };

  var runner = new KodeRunr($("#ext").val());

  $("#submitCode").on("click", runner.runCode.bind(runner));

  $("#ext").on("change", function() {
    // Empty the screen
    runner.editor.setValue("", undefined);
    $("#streamingResult").text("");
    console.log(this.value);
    runner.setExt(this.value);
  })
});
