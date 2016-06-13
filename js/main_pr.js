(function(){
    var MsgBox = function(title, msg, inhtml){

      //初始化参数
      this.width = "500px";
      this.height = "300px";
      this.title = title;
      this.msg = msg;
      this.i = 0;
      this.inhtml = inhtml;

      this.init = function(){

        this.i ++;

        var html = "<div class='commonmsgbox'><div class='title'>";
        html += this.title;
        html += "</div><div class='msgclose'>关闭</div><div class='commonmsg'>";
        html += this.msg;
        html += "</div><div class='content'>";  
        html += this.inhtml;
        html += "</div></div>";

        if(this.i == 1) $(".wrapper").append(html);
        else $(".commonmsgbox").replaceWith(html);

        $(".commonmsgbox").css({width:this.width,height:this.height,display:"none"});

        var msgtop = (document.documentElement.clientHeight - parseInt($(".commonmsgbox").css("height")))/2;
        var msgleft = (document.documentElement.clientWidth > 1003) ? (document.documentElement.clientWidth - parseInt($(".commonmsgbox").css("width"))) / 2 : (1003 - parseInt($(".commonmsgbox").css("width"))) / 2;

        $(".commonmsgbox").css({"top":msgtop + "px","left":msgleft + "px"});
        $(".commonmsgbox .title").css("width",$(".commonmsgbox").css("width"));

        var _this = this;

        $(".commonmsgbox .msgclose").click(function(){
            _this.hide();
        });
      };

      this.show = function(){
        $(".commonmsgbox").slideDown("fast");
      };

      this.hide = function(){
        this.inhtml = "";
        $(".commonmsgbox").html("");
        $(".commonmsgbox").slideUp("fast");
      };

      this.close = this.hide;
    };

    var msg = new MsgBox("AlloyPhoto","","");

    

    //对滑动bar的事件处理对象
    var Bar = {
        observer: [],

        notify: function(target){
            var value = parseInt(target.parent().find(".dMsg").text());

            for(var i = 0; i < this.observer.length; i ++){

               if(target.parent().attr("id") == this.observer[i][0]){
                    this.observer[i][1](value);
               }

            }
        },

        addObserver: function(targetId, func){
            this.observer.push([targetId, func]);
        }
    };

    var COM_MODEL = [
        "正常", "颜色减淡", "变暗", "变亮", "正片叠底", "滤色", "叠加", "强光", "差值", "排除", "点光", "颜色加深", "线性加深", "线性减淡", "柔光", "亮光", "线性光", "实色混合"
    ];

    var COM_HTML_MODEL = "<select class='commodel'>";

    for(var i = 0; i < COM_MODEL.length; i ++){
        COM_HTML_MODEL += "<option value='" + COM_MODEL[i] + "'>" + COM_MODEL[i] + "</option>";
    }

    COM_HTML_MODEL += "</select>";

    //图层的命名计数
    var layerCount = 0;
    var Main = {
        layers: [],

        //主画布
        ps: null,

        layers: [],

        currLayer: [],
        args:{},
        //+openFile
        //打开文件
        openFile: function(fileUrl){
            //+0528添加
            var xmlHttpRequest = new XMLHttpRequest();
            xmlHttpRequest.open("POST", '/upload', true);
            var formData = new FormData();
            // This should automatically set the file name and type.
            formData.append("upload", fileUrl);
            // Sending FormData automatically sets the Content-Type header to multipart/form-data
            xmlHttpRequest.send(formData);
            //-0528添加




            var reader = new FileReader();
            var _this = this;

            reader.readAsDataURL(fileUrl);

            reader.onload = function(){
                msg.close();

                var img = new Image();
                img.src = this.result;
                img.onload = function(){
                    _this.addImage(img);
                };

            };
        
        },
        //-openFile
        
        //+addImage
        addImage: function(img){

            var psObj = AlloyImage(img);

            if(!(this.ps)){
                this.ps = AlloyImage(parseInt(img.width),parseInt(img.height),"rgba(255,255,255,0)");

                $(".right").css({width:img.width,height:img.height});
                $(".openFile").html("画布区");
            }

            this.layers.push(psObj);

            //添加一个图层
            this.ps.addLayer(psObj);

            //设置当前图层
            this.currLayer = [this.layers.length - 1];
            
            //向面板添加一个图层
            //this.addLayer();
            this.draw();
        },
        //-addImage
        
        //+receiveImg
        //监听 Q+ rpc传来的文件做处理
        recieveImg:function(){
            var _this = this;
            /*
           RpcMgr.recive( 'fdeee', function( data ){

                    if ( data.music_name === 'mm' ){
                            this.notice({"msg": data.music_name + " succ"});
                    }else{
                            this.notice({"msg": data.music_name + " fail"});
                    }

                    _this.notifer = this;
                  var imgUrl = data.imgUrl;

                  if(imgUrl){
                    var img = new Image();
                    img.src = imgUrl;

                    img.onload = function(){
                        _this.addImage(img);
                    };
                  }

                                                
            });
            */
        },
        //-receiveImg
        
        //+attachE
        attachE: function () {
            /*
             * @description:事件处理
             *
             * */

            var _this = this;

            //上传文件处理
            $("#upFile").change(function (e) {
                msg.title = "读取文件";
                msg.msg = "正在读取文件...";
                msg.inhtml = "<img src='style/image/03.gif' />";
                msg.init();
                msg.show();

                _this.openFile(e.target.files[0]);
            });
            //弹出模糊类型和模糊半径设置框
            $("#new").click(function () {
                msg.title = "新建";
                msg.msg = "新建立一个模糊核";
                msg.inhtml="<input type='radio' name='mode' value='gaussian' />高斯模糊<input type='radio' name='mode' value='motion' checked='checked'/>运动模糊<br />";
                msg.inhtml+="<input type='radio' name='method' value='wiener' checked='checked'/>维纳滤波<input type='radio' name='method' value='leastsquare'/>最小二乘方滤波<br />";
                
                msg.inhtml += "半径：<input type='number' id='newWidth' value='10' />px<input type='button' id='confirmNew' value='确定' />";
                msg.init();
                msg.show();
            });
            //设置模糊类型和模糊半径
            $("#confirmNew").live("click", function () {
                _this.args.mode=$("input[name='mode']:checked").val();
                _this.args.method=$("input[name='method']:checked").val();
                _this.args.radius=$("#newWidth").val();
                msg.hide();
            });

            $("#modify").click(function () {

                $(".pItem").hide("fast");
                $("#upFile").hide();
                $(".modifyItem").css("display", "block");
                $(".back").show();

            });

            $("#lj").click(function () {



                $(".ljItem").css("display", "block");


                $("#upFile").hide();
                $(".back").show();
            });

            $("#saveFile").click(function () {
                var data = _this.ps.save();
                //var img = new Image();
                //img.src = data;
                //$(".painting").html(img);
                alert("图片已经输出成功，请右键点击图片，另存为图片即可保存");
                /*
                try{
                    _this.notifer.notice({"msg": data});
                }catch(e){
                    alert(e.message);
                }
                */
            });

            $(".ljItem").click(function () {
                var text = $(this).text();

                for (var i = 0; i < _this.currLayer.length; i++) {
                    _this.layers[_this.currLayer[i]].act(text);
                }

                _this.draw();
            });

            $(".back").click(function () {
                $(".pItem").show("fast");
                $("#upFile").show();
                $(".subItem").hide();
                $(this).hide();
            });
            $("#modi_b").click(function () {

                msg.title = "模糊方向调节";
                msg.msg = "请滑动调节";
                msg.inhtml = "模糊方向:<div id='dBar1' class='dBar' rangeMin='-50' rangeMax='50'><a draggable='false' href='#'></a><div class='dMsg'>0</div></div><br />";
                msg.inhtml += "<div class='dView'><button id='excute'>确定</button><button id='cancel'>取消</button></div>";
                msg.init();
                msg.show();

                // Bar.addObserver("dBar1", function (value) {
                //     var value2 = parseInt($("#dBar2 .dMsg").text());
                //     for (var i = 0; i < _this.currLayer.length; i++) {
                //         _this.layers[_this.currLayer[i]].view("亮度", value, value2);
                //     }
                //     _this.draw();
                // });
                
                Bar.addObserver("dBar1", function (value) {
                    
                    _this.args.direction=value;
                });
            });

            $("#modi_HSI").click(function () {

                msg.title = "sigma调节";
                msg.msg = "请滑动调节";
                msg.inhtml = "sigma:<div id='dBar2' class='dBar' rangeMin='-180' rangeMax='180'><a draggable='false' href='#'></a><div class='dMsg'>0</div></div><br />";
                msg.inhtml += "<div class='dView'><button id='excute'>确定</button><button id='cancel'>取消</button></div>";
                msg.init();
                msg.show();

                Bar.addObserver("dBar2", function (value) {
                    _this.args.sigma=value;
                
                });
            });


            $("#excute").live("click", function () {
                //加一点效果在这里比较好
                msg.hide();
            });

            $("#cancel").live("click", function () {

                msg.hide();
            });
            
            //开始处理了
            $("#processing").click(function () {
                console.log(_this.args.mode);
                var xmlHttpRequest = new XMLHttpRequest();
                xmlHttpRequest.open("POST", '/show', true);
                xmlHttpRequest.setRequestHeader("Content-Type","application/x-www-form-urlencoded");
                xmlHttpRequest.onreadystatechange=processResponse
                var formData = new FormData();
                if (_this.args.mode=="gaussian") {
                    strData="mode="+_this.args.mode;
                    strData+="&method="+_this.args.method;
                    strData+="&radius="+_this.args.radius;
                    strData+="&sigma="+_this.args.sigma;
                    strData+="&direction=0.0";
                   
                    console.log(strData);
                    xmlHttpRequest.send(strData);
                } else {
                    // formData.append("mode", "motion");
                    // formData.append("method","wiener");
                    // formData.append("radius","20");
                    // formData.append("direction","-0.33");
                    strData="mode="+_this.args.mode;
                    strData+="&method="+_this.args.method;
                    strData+="&radius="+_this.args.radius;
                    strData+="&sigma=0.0";
                    strData+="&direction="+_this.args.direction;
                   console.log(strData);
                    xmlHttpRequest.send(strData);
                }
                //xmlHttpRequest.send(formData);
                
                function processResponse(){
                    if(xmlHttpRequest.readyState==4){ //判断对象状态4代表完成
                        if(xmlHttpRequest.status==200){ //信息已经成功返回，开始处理信息
                            var strObj=eval("("+xmlHttpRequest.responseText+")");
                            console.log(strObj.Dataurl)
                            var img=$("<img>");
                            img.attr("src","data:image/jpg;base64,"+strObj.Dataurl);
                            $(".painting").append(img);
                            $(".painting canvas").remove();
                        }
                    }
                }
                
            });

            
            $(".msgclose").live("click", function () {
                for (var i = 0; i < _this.currLayer.length; i++) {
                    _this.layers[_this.currLayer[i]].cancel();
                }

                _this.draw();
                msg.hide();
            });

            var flagM1 = 0, flagD1 = 0;//滑动bar标记
            var orginOffsetX = 0;
            var clientX = 0, offsetX = 0, dTarget;

            $(".dBar a").live("mousedown", function (e) {
                flagM1 = 1;
                clientX = e.clientX;
                offsetX = parseInt($(this).css("left"));
                dTarget = $(this);
            });

            $(document).bind("mousemove", function (e) {
                if (flagM1) {
                
                    //拖拽开始
                    flagD1 = 1;
                    var dx = e.clientX - clientX;
                    var currLeft = offsetX + dx;
                    var circleWidth = parseInt(dTarget.css("width")) / 2;
                    var parentWidth = parseInt(dTarget.parent().css("width")) - circleWidth;

                    if (currLeft > parentWidth) currLeft = parentWidth;
                    if (currLeft < -circleWidth) currLeft = - circleWidth;
                    dTarget.css("left", currLeft + "px");

                    var rangeMin = parseFloat(dTarget.parent().attr("rangeMin")) || 0;
                    var rangeMax = parseFloat(dTarget.parent().attr("rangeMax")) || 0;
                    var percent = (currLeft + circleWidth) / parentWidth;
                    var nowRange = (rangeMin + (rangeMax - rangeMin) * percent).toFixed(0);

                    dTarget.parent().find(".dMsg").text(nowRange);
                }
            });


            //混合模式改变时处理程序
            $(".commodel").live("change", function () {
                var comModel = $(this).val();
                for (var i = 0; i < _this.currLayer.length; i++) {
                    _this.ps.layers[_this.currLayer[i]][1] = comModel;
                }
                _this.draw();
            });

            var flagK = 0, flagM = 0, flagD = 0;//flagK 标记key alt flagM标记 Mouse flagD标记拖拽事件
            var dx = [], dy = [];

            $(document).keydown(function (e) {
                if (e.keyCode == 17) {
                    flagK = 1;
                    $(".left").css("cursor", "move");
                }
            }).keyup(function (e) {
                if (e.keyCode == 17) {
                    flagK = 0;
                    $(".left").css("cursor", "auto");
                }
            }).mouseup(function (e) {
                flagM = 0;
                flagM1 = 0;
                if (flagD) {
                    _this.draw();
                    flagD = 0;//标记拖拽结束
                }
                if (flagD1) {//滑动bar拖拽结束
                    Bar.notify(dTarget);
                    flagD1 = 0;
                }
            });
            $(".painting").get(0).onmousedown = function(e){
                var offsetX = e.offsetX ? e.offsetX : e.layerX;
                    var offsetY = e.offsetY ? e.offsetY : e.layerY;

                    for(var i = 0;i < _this.currLayer.length;i ++){
                        var lDx = _this.ps.layers[_this.currLayer[i]][2] || 0;
                        var lDy = _this.ps.layers[_this.currLayer[i]][3] || 0;
                        dx[i] = offsetX - lDx; 
                        dy[i] = offsetY - lDy; 
                    }

                    flagM = 1;

            };
            
            $(".painting").get(0).onmousemove = function(e){
                    if(flagK && flagM){
                        var offsetX = e.offsetX ? e.offsetX : e.layerX;
                        var offsetY = e.offsetY ? e.offsetY : e.layerY;

                        for(var i = 0;i < _this.currLayer.length;i ++){
                            _this.ps.layers[_this.currLayer[i]][2] = offsetX - dx[i];
                            _this.ps.layers[_this.currLayer[i]][3] = offsetY - dy[i];
                        }
                        _this.draw(true);
                        flagD = 1;//标记拖拽发生
                    }
            };

        },
        //-attachE
        
        //+draw
        draw: function(isFast){
                //显示主画布
                this.ps.show(".painting",isFast);

                //重绘直方图
                this.ps.drawRect();
        },
        //-draw
        //+addLayer
        addLayer: function(){
           var html = "<div class='lItem'><span class='layerName'>图层" + (++ layerCount) + "</span> 混合模式" + COM_HTML_MODEL + "</div>"; 
           $(".layer").prepend(html);

           this.showCurrLayer();
        },
        //-addLayer
        //+addInputFile
        addInputFile: function(){
            var html = "<input type='file' id='upFile' style='z-index:1'/>";
            //$(".panel").append(html);
            $("#effects").prepend(html);
        },
        //-addInputFile
        
        //+init
        init: function(){
            this.addInputFile();
            this.recieveImg();
            this.attachE();
        },
        //-init
    };

    Main.init();

})();
