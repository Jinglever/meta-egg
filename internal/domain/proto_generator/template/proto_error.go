package template

var TplProtoError string = `syntax = "proto3";
package %%PROJECT-NAME-PKG%%;
option go_package = "./%%PROJECT-NAME-DIR%%";

enum ErrCode {
    // OK is returned on success.
	Ok = 0; // 成功

    // Unknown error. An example of where this error may be returned is
	// if a Status value received from another address space belongs to
	// an error-space that is not known in this address space. Also
	// errors raised by APIs that do not return enough error information
	// may be converted to this error.
	Unknown = 2; // 未知错误

    // InvalidArgument indicates client specified an invalid argument.
	// Note that this differs from FailedPrecondition. It indicates arguments
	// that are problematic regardless of the state of the system
	// (e.g., a malformed file name).
    InvalidArgument = 3; // 参数错

    // NotFound means some requested entity (e.g., file or directory) was
	// not found.
    NotFound = 5; // 实体不存在

    // AlreadyExists means an attempt to create an entity failed because one
	// already exists.
    AlreadyExists = 6; // 创建实体时冲突

    // PermissionDenied indicates the caller does not have permission to
	// execute the specified operation. It must not be used for rejections
	// caused by exhausting some resource (use ResourceExhausted
	// instead for those errors). It must not be
	// used if the caller cannot be identified (use Unauthenticated
	// instead for those errors).
    PermissionDenied = 7; // 权限不足

    // ResourceExhausted indicates some resource has been exhausted, perhaps
	// a per-user quota, or perhaps the entire file system is out of space.
    ResourceExhausted = 8; // 资源不足

    // FailedPrecondition indicates operation was rejected because the
	// system is not in a state required for the operation's execution.
	// For example, directory to be deleted may be non-empty, an rmdir
	// operation is applied to a non-directory, etc.
	//
	// A litmus test that may help a service implementor in deciding
	// between FailedPrecondition, Aborted, and Unavailable:
	//  (a) Use Unavailable if the client can retry just the failing call.
	//  (b) Use Aborted if the client should retry at a higher-level
	//      (e.g., restarting a read-modify-write sequence).
	//  (c) Use FailedPrecondition if the client should not retry until
	//      the system state has been explicitly fixed. E.g., if an "rmdir"
	//      fails because the directory is non-empty, FailedPrecondition
	//      should be returned since the client should not retry unless
	//      they have first fixed up the directory by deleting files from it.
	//  (d) Use FailedPrecondition if the client performs conditional
	//      REST Get/Update/Delete on a resource and the resource on the
	//      server does not match the condition. E.g., conflicting
	//      read-modify-write on the same resource.
    FailedPrecondition = 9; // 前置条件失败

    // Aborted indicates the operation was aborted, typically due to a
	// concurrency issue like sequencer check failures, transaction aborts,
	// etc.
	//
	// See litmus test above for deciding between FailedPrecondition,
	// Aborted, and Unavailable.
    Aborted = 10; // 操作被中止

    // OutOfRange means operation was attempted past the valid range.
	// E.g., seeking or reading past end of file.
	//
	// Unlike InvalidArgument, this error indicates a problem that may
	// be fixed if the system state changes. For example, a 32-bit file
	// system will generate InvalidArgument if asked to read at an
	// offset that is not in the range [0,2^32-1], but it will generate
	// OutOfRange if asked to read from an offset past the current
	// file size.
	//
	// There is a fair bit of overlap between FailedPrecondition and
	// OutOfRange. We recommend using OutOfRange (the more specific
	// error) when it applies so that callers who are iterating through
	// a space can easily look for an OutOfRange error to detect when
	// they are done.
    OutOfRange = 11; // 超出范围

    // Internal errors. Means some invariants expected by underlying
	// system has been broken. If you see one of these errors,
	// something is very broken.
    Internal = 13; // 内部错误

    // Unavailable indicates the service is currently unavailable.
	// This is a most likely a transient condition and may be corrected
	// by retrying with a backoff. Note that it is not always safe to retry
	// non-idempotent operations.
	//
	// See litmus test above for deciding between FailedPrecondition,
	// Aborted, and Unavailable.
    Unavailable = 14; // 服务不可用，请重试

    // DataLoss indicates unrecoverable data loss or corruption.
    DataLoss = 15; // 数据丢失或损坏

    // Unauthenticated indicates the request does not have valid
	// authentication credentials for the operation.
    Unauthenticated = 16; // 未认证，客户端未提供凭据或提供的凭据无效

	MaxReservedErrCode = 1000; // 预留错误码最大值
    //////////////// custom err code start from 1001 //////////////
}
`
