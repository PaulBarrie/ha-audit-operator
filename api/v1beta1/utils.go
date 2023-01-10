package v1beta1

type NamespacedName struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}
