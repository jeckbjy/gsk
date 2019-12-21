# filter
这里实现一些常用的Filter插件,这里会依赖其他模块,比如FrameFilter,PacketFilter,ExecutorFilter

## 常见的处理流程
InBound(Read)   ===> TransferFilter ===> FrameFilter ===> PacketFilter ===> ExecutorFilter
OutBound(Write) <=== TransferFilter <=== FrameFilter <=== PacketFilter