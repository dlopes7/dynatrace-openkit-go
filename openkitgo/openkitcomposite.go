package openkitgo

const DEFAULT_ACTION_ID = 0

type OpenKitComposite interface {
	storeChildInList(o OpenKitObject)
	removeChildFromList(o OpenKitObject) bool
	getCopyOfChildObjects() []OpenKitObject
	getChildCount() int
	onChildClosed(child OpenKitObject)
	getActionID() int
	close()
}

type OpenKitObject interface {
	close()
}
