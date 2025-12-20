package handle

import (
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	// SystemHandleInformation is the information class for querying system handles (32-bit)
	SystemHandleInformation = 16

	// SystemExtendedHandleInformation is the information class for querying system handles (64-bit)
	SystemExtendedHandleInformation = 64

	// ObjectNameInformation is the information class for querying object names
	ObjectNameInformation = 1

	// ObjectTypeInformation is the information class for querying object types
	ObjectTypeInformation = 2

	// StatusInfoLengthMismatch indicates buffer is too small
	StatusInfoLengthMismatch = 0xC0000004

	// DuplicateCloseSource closes the source handle
	DuplicateCloseSource = 0x00000001

	// DuplicateSameAccess duplicates handle with same access
	DuplicateSameAccess = 0x00000002
)

var (
	ntdll                 = windows.NewLazySystemDLL("ntdll.dll")
	procNtQuerySystemInfo = ntdll.NewProc("NtQuerySystemInformation")
	procNtQueryObject     = ntdll.NewProc("NtQueryObject")
	procNtDuplicateObject = ntdll.NewProc("NtDuplicateObject")
)

// SystemHandleTableEntryInfo represents a single handle entry in the system (32-bit)
type SystemHandleTableEntryInfo struct {
	UniqueProcessID       uint16
	CreatorBackTraceIndex uint16
	ObjectTypeIndex       byte
	HandleAttributes      byte
	HandleValue           uint16
	Object                uintptr
	GrantedAccess         uint32
}

// SystemHandleInformationEx represents system handle information (32-bit)
type SystemHandleInformationEx struct {
	NumberOfHandles uintptr
	Reserved        uintptr
	Handles         [1]SystemHandleTableEntryInfo
}

// SystemHandleTableEntryInfoEx represents a single handle entry in the system (64-bit)
type SystemHandleTableEntryInfoEx struct {
	Object                uintptr
	UniqueProcessID       uintptr
	HandleValue           uintptr
	GrantedAccess         uint32
	CreatorBackTraceIndex uint16
	ObjectTypeIndex       uint16
	HandleAttributes      uint32
	Reserved              uint32
}

// SystemExtendedHandleInformationEx represents system handle information (64-bit)
type SystemExtendedHandleInformationEx struct {
	NumberOfHandles uintptr
	Reserved        uintptr
	Handles         [1]SystemHandleTableEntryInfoEx
}

// UnicodeString represents a Windows UNICODE_STRING
type UnicodeString struct {
	Length        uint16
	MaximumLength uint16
	Buffer        *uint16
}

// ObjectNameInfo represents object name information
type ObjectNameInfo struct {
	Name UnicodeString
}

// ObjectTypeInfo represents object type information
type ObjectTypeInfo struct {
	TypeName UnicodeString
}

// ntQuerySystemInformation queries system information
func ntQuerySystemInformation(
	systemInformationClass uint32,
	systemInformation uintptr,
	systemInformationLength uint32,
	returnLength *uint32,
) (ntstatus error) {
	r0, _, _ := syscall.SyscallN(
		procNtQuerySystemInfo.Addr(),
		uintptr(systemInformationClass),
		systemInformation,
		uintptr(systemInformationLength),
		uintptr(unsafe.Pointer(returnLength)),
	)
	if r0 != 0 {
		ntstatus = syscall.Errno(r0)
	}
	return
}

// ntQueryObject queries object information
func ntQueryObject(
	handle windows.Handle,
	objectInformationClass uint32,
	objectInformation uintptr,
	objectInformationLength uint32,
	returnLength *uint32,
) (ntstatus error) {
	r0, _, _ := syscall.SyscallN(
		procNtQueryObject.Addr(),
		uintptr(handle),
		uintptr(objectInformationClass),
		objectInformation,
		uintptr(objectInformationLength),
		uintptr(unsafe.Pointer(returnLength)),
	)
	if r0 != 0 {
		ntstatus = syscall.Errno(r0)
	}
	return
}

// ntDuplicateObject duplicates an object handle
func ntDuplicateObject(
	sourceProcessHandle windows.Handle,
	sourceHandle windows.Handle,
	targetProcessHandle windows.Handle,
	targetHandle *windows.Handle,
	desiredAccess uint32,
	handleAttributes uint32,
	options uint32,
) (ntstatus error) {
	r0, _, _ := syscall.SyscallN(
		procNtDuplicateObject.Addr(),
		uintptr(sourceProcessHandle),
		uintptr(sourceHandle),
		uintptr(targetProcessHandle),
		uintptr(unsafe.Pointer(targetHandle)),
		uintptr(desiredAccess),
		uintptr(handleAttributes),
		uintptr(options),
	)
	if r0 != 0 {
		ntstatus = syscall.Errno(r0)
	}
	return
}

// getUnicodeString converts a UnicodeString to a Go string
func getUnicodeString(us *UnicodeString) string {
	if us.Buffer == nil || us.Length == 0 {
		return ""
	}

	// Create a slice from the buffer
	length := us.Length / 2 // UTF-16 is 2 bytes per character
	slice := unsafe.Slice(us.Buffer, length)

	return windows.UTF16ToString(slice)
}
