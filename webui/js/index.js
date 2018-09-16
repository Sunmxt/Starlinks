
document.addEventListener("DOMContentLoaded", function() {
    var req = null;

    url_input = new Vue({
        el: "#LongURL"
        , data: {
            TargetURL: ""
        }
    })

    prompt_box = new Vue({
        el: "#ResultBox"
        , data: {
            message: "成功添加短链接：http://l.stuhome.com/E1KmC2d98A6"
            , show: false
            , prompt_class: "alert alert-success"
        }
    })

    clear_btn = new Vue({
        el: "#clear-url"
        , methods: {
            ClearURL: function () {
                url_input.TargetURL = ""
                prompt_box.show=false
            }
        }
    })

    append_btn = new Vue({
        el: "#append-url"
        , methods: {
            AppendURL: function (){
                if(url_input.TargetURL == "") {
                    prompt_box.prompt_class = "alert alert-warning"
                    prompt_box.message = "链接长度过短."
                    prompt_box.show = true
                    return
                }

                if(req != null) {
                    req.abort()
                    req = null
                }

                req = new XMLHttpRequest();
                req.onreadystatechange = function () {
                    if(req.readyState == 4) {
                        if(req.status == 200){
                            try {
                                result = JSON.parse(req.responseText)
                                if(result.ErrorCode != 0) {
                                    prompt_box.prompt_class = "alert alert-danger"
                                    prompt_box.message = "添加短链接失败 (" + result.ErrorCode + "):" + result.ErrorDesc
                                    prompt_box.show = true
                                } else {
                                    prompt_box.prompt_class = "alert alert-success"
                                    prompt_box.message = "成功添加短链接:"+ window.location.protocol + "//" + window.location.host + result.ShortRoute
                                    prompt_box.show = true
                                }
                            } catch(e) {
                                console.log("Invalid server response: ", req.responseText)
                                throw e
                            }
                        }
                        req = null
                    }
                }
                req.ontimeout = function() {
                    prompt_box.prompt_class = "alert alert-warning"
                    prompt_box.message = "添加短链接超时"
                    prompt_box.show = true
                    req = null
                }

                req.open("POST", "/v1/links")
                var args = "URL=" + url_input.TargetURL
                req.setRequestHeader('Content-type', 'application/x-www-form-urlencoded');
                req.send(args)
            }
        }
    })
})
