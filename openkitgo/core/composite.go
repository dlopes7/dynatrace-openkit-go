package core

const (
	DEFAULT_ACTION_ID = 0
)

type OpenKitComposite interface {
	StoreChildInList(child OpenKitObject)
	RemoveChildFromList(child OpenKitObject) bool
	GetCopyOfChildObjects() []OpenKitObject
	GetChildCount() int
	OnChildClosed(child OpenKitObject)
	GetActionID() int
}
