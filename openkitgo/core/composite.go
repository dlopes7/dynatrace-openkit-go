package core

const (
	DEFAULT_ACTION_ID = 0
)

type OpenKitComposite interface {
	storeChildInList(child OpenKitObject)
	removeChildFromList(child OpenKitObject) bool
	getCopyOfChildObjects() []OpenKitObject
	getChildCount() int
	onChildClosed(child OpenKitObject)
	getActionID() int
}
