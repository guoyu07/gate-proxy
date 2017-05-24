package gateway

import "goodsogood/errors"

var (
    BackendServiceNotAvailable = errors.New(-9001, "没有可用的后端服务")
    BackendAlreadyExist = errors.New(-9002, "后端服务已经存在")
    BackendNotFound = errors.New(-9009, "后端服务不存在")
    BackendServiceError = errors.New(-9003, "服务不可用")
    // Attr Is Required -9004
    APINotFound = errors.New(-9005, "请求的接口不存在")
    ClusterAlreadyExist = errors.New(-9006, "集群已经存在")
    ClusterNotFound = errors.New(-9007, "集群不存在")
    APIAlreadyExist = errors.New(-9008, "API接口已经存在")

    PluginAlreadyExist = errors.New(-9010, "插件已经存在")

    ParamParseFailed = errors.New(-9011, "参数解析失败")
    ClusterNameEmpty = errors.New(-9012, "Cluster 名称不能为空")
    SchemaUnknowable = errors.New(-9013, "Schema 不能识别")
    AddrUnknowable = errors.New(-9014, "Addr 不能识别")
    HeartPathNotEmpty = errors.New(-9015, "心跳地址不能为空")
    MaxQPSNotZero = errors.New(-9016, "最大QPS不能为0")
    UnknowableMethod = errors.New(-9017, "无法识别的Method")
    SUCCESS = errors.New(0, "操作成功")
)

