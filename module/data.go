package module

import "net/http"

// 数据的接口类型
type Data interface {
	//判断数据是否有效
	Valid() bool
}


// 数据请求结构
type Request struct {
	// HTTP请求
	httpReq *http.Request
	// 请求深度
	depth uint32
}

// 创建请求
func NewRequest(httpReq *http.Request, depth uint32) *Request {
	return &Request{
		httpReq: httpReq,
		depth:   depth,
	}
}

// 获取请求
func (req Request) HTTPReq() *http.Request {
	return req.httpReq
}

// 获取请求深度
func (req Request) Depth() uint32  {
	return req.depth
}

// 判断请求是否有效
func (req *Request) Valid() bool  {
	return req != nil && req.httpReq.URL != nil
}


// 数据响应结构
type Response struct {
	// HTTP响应
	httpResp *http.Response
	// 响应深度
	depth uint32
}

// 创建响应
func NewRespone(httpResp *http.Response, depth uint32) *Response  {
	return &Response{
		httpResp: httpResp,
		depth:    depth,
	}
}

// 获取响应
func (resp Response) HTTPResp() *http.Response  {
	return resp.httpResp
}

// 获取响应深度
func (resp Response) Depth() uint32  {
	return resp.depth
}

// 判断响应是否有效
func (resp *Response) Valid() bool  {
	return resp.httpResp != nil && resp.httpResp.Body != nil
}


// 条目类型
type Item map[string]interface{}

// 判断条目是否有效
func (item Item) Valid() bool  {
	return item != nil
}