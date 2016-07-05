Configuration
======

应用程序的配置控制

#### 配置方式 ####
支持以下三种方式

- ini配置文件(默认文件名./config.ini)
- etcd服务
- 纯环境变量

文件和etcd配置项也也会包含系统环境变量包含环境变量

配置方式通过环境变量 **GLOBAL_CONF** 来控制，以上两种方式分别对应

- file::/root/ect/wx.conf
- etcd::http://192.168.10.7:2379
- env:://

环境变量GLOBAL_CONF的缺省值为 file::./config.ini

#### 配置key ####

配置key使用点号划分段，例如

- wmds.mp.appid
- wx.oracle.host

key大小写不明感，建议使用小写字母,针对etcd的存值，一定要全部小写

#### 程序变量反射值 ####

    type struct Config {
        OracleHost string `conf:"wx.oracle.host,omit"`
        OracleHostDef string `conf:"wx.oracle.host,default(sss)"`
        InlineStruct struct{
            StringsValue []string `conf:"strings"
        } `conf:"com.struct"`
    }
    
omit标明，可缺省。否则值必须传入
default为默认值，如果配置没有值则取值为默认值,数组的默认值以；分割default(1;2;3)；map则对应：分割key value default(key1:value1;key2:value2)
支持标签叠加,StringsValue的最终标签为com.struct.strings

通过上述变量的标注取值

#### 注释 ####

文件配置方式使用，行开始#注释

#### 环境变量默认值 ####

对应配置文件或etcd中不存在的配置项，会取环境变量的值来替换

如果文件配置的一行为[**]，则内容忽略

#### 支持的基本取值类型 ####
 
 - 字符串类型
 - bool类型，只有设置值为1，T，t，true,TRUE,True时为true,其他为false
 - int类型
 - Float32类型
 - Float64类型
 - 以上基本类型的切片类型，切片的设置值为"；"号分割

struct field类型
 
 - 以上基本类型，及其切片类型
 - map[string]string 类型
 - []struct 类型
 - map[string]struct 类型

####  map解析key中包含.号需要做特殊处理 ###

例如配置项

    wmds.wxoauth2.redirects."/menu_usercenter.html" = /usercenter.html
    wmds.wxoauth2.redirects."/menu_vipcard.html" = "/vipcard.html"
    wmds.wxoauth2.redirects."/menu_index.html" = "/index.html"
    wmds.wxoauth2.redirects."/menu_login.html" = "/login.html"
    wmds.wxoauth2.redirects."/menu_bonusgot.html" = "/bonusgot.html"
    wmds.wxoauth2.redirects.unauthorized = "/login.html"
    
解析值为

        {
                "/menu_usercenter.html":"/usercenter.html",
                "/menu_vipcard.html":"/vipcard.html",
                "/menu_index.html":"/index.html",
                "/menu_login.html":"/login.html",
                "/menu_bonusgot.html":"/bonusgot.html",
                "unauthorized":"/login.html"
            }
