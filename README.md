# CoolQ-MCList
一款Go语言编写的酷Q插件，可以批量MCPing服务器
请在release内下载cpk文件
使用/list来Ping服务器
不可用于任何商业用途
**使用说明：**
MCList使用说明：
请遵循以下格式进行config.json文件的书写
{
    "des":"欢迎来到XXX服务器",
    "servers":[
        {
            "port":"端口",
            "ip":"IP1",
            "name":"简称"
        },
        
        {
            "port":"端口",
            "ip":"IP2",
            "name":"简称"
        }
    ]
}

以此类推，可以加入任意数量服务器
*config.json文件实时读取，修改完毕可直接使用，无需重启插件
*config.json文件中每行字符数量不超过65536个
