namespace java com.goodsogood.service.union.auth.thrift
namespace go goodsogood.thrift.protocol.token.service

include "/Users/liaocaixin/Projects/goodsogood/goodsogood_union_thrift/base/thrift/RID.thrift"
include "TokenResp.thrift"
include "/Users/liaocaixin/Projects/goodsogood/goodsogood_union_thrift/base/thrift/Resp.thrift"
include "/Users/liaocaixin/Projects/goodsogood/goodsogood_union_thrift/words/thrift/Word.thrift"

/*
** token验证、生成服务
*/
service TokenService{
 
     /**
     ** 创建访问token
     ** @param rId 请求ID
     ** @param userId 用户ID
     ** @param plateformDeviceType 设备类型
     ** @param plateformDeviceInfo 设备ID
     **/
     TokenResp.TokenResp createAccessToken(1:required RID.RID rId, 
                                           2:required string userId, 
                                           3:required string plateformDeviceType,
                                           4:required string plateformDeviceInfo)
 
 
     /**
     ** 检测访问token是否有效
     ** @param rId 请求ID
     ** @param userId 用户ID
     ** @param plateformDeviceType 设备类型
     ** @param plateformDeviceInfo 设备ID
     ** @param token 待检测的token
     **/
     Resp.Resp  checkAccessToken(1:required RID.RID rId, 
                              2:required string userId, 
                              3:required string plateformDeviceType, 
                              4:required string plateformDeviceInfo,
                              5:required string token)

    /**
    * 检测服务可用性
    **/
    Word.ServiceState ping()
}