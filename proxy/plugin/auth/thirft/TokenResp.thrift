namespace java com.goodsogood.service.union.auth.thrift
namespace go goodsogood.thrift.protocol.token

/**
** token生成的返回值
**/
struct TokenResp{
    1:required  string   rId,
    2:required  i32      code,         // 返回状态码
    3:required  i64      timestamp,    // 服务器时间
    4:string   message,    //
    5:string   token,
}