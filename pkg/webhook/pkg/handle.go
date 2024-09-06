package pkg

import (
	"context"
	"fmt"
	"net/http"
)

func (h *MutatingHandler) Handle(ctx context.Context, req kubeadmission.Request) kubeadmission.Response {
	reqJson, err := json.Marshal(req.AdmissionRequest)
	if err != nil {
		return kubeadmission.Errored(http.StatusBadRequest, err)
	}
	fmt.Println(string(reqJson))
	obj := &appsv1.Deployment{}

	err = h.Decoder.Decode(req, obj)
	if err != nil {
		return kubeadmission.Errored(http.StatusBadRequest, err)
	}
	fmt.Println(obj.Name)
	originObj, err := json.Marshal(obj)
	if err != nil {
		return kubeadmission.Errored(http.StatusBadRequest, err)
	}
	// 将新的资源副本数量改为1
	newobj := obj.DeepCopy()
	replicas := int32(1)
	newobj.Spec.Replicas = &replicas
	currentObj, err := json.Marshal(newobj)
	if err != nil {
		return kubeadmission.Errored(http.StatusBadRequest, err)
	}
	// 对比之前的资源类型和之后的资源类型的差异生成返回数据
	resp := kubeadmission.PatchResponseFromRaw(originObj, currentObj)
	respJson, err := json.Marshal(resp.AdmissionResponse)
	if err != nil {
		return kubeadmission.Errored(http.StatusBadRequest, err)
	}
	return resp
}
