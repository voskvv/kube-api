package handlers

import (
	"net/http"

	"git.containerum.net/ch/kube-api/pkg/kubernetes"
	"git.containerum.net/ch/kube-api/pkg/model"
	m "git.containerum.net/ch/kube-api/pkg/router/midlleware"
	cherry "git.containerum.net/ch/kube-client/pkg/cherry/kube-api"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	log "github.com/sirupsen/logrus"
)

const (
	secretParam = "secret"
)

func GetSecretList(ctx *gin.Context) {
	log.WithFields(log.Fields{
		"Namespace Param": ctx.Param(namespaceParam),
		"Namespace":       ctx.MustGet(m.NamespaceKey).(string),
	}).Debug("Get secret list Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	secrets, err := kube.GetSecretList(ctx.MustGet(m.NamespaceKey).(string))
	if err != nil {
		ctx.Error(err)
		cherry.ErrUnableGetResourcesList().Gonic(ctx)
		return
	}

	ret, err := model.ParseSecretList(secrets)
	if err != nil {
		ctx.Error(err)
		cherry.ErrUnableGetResourcesList().Gonic(ctx)
		return
	}

	ctx.JSON(http.StatusOK, ret)
}

func GetSecret(ctx *gin.Context) {
	log.WithFields(log.Fields{
		"Namespace Param": ctx.Param(namespaceParam),
		"Namespace":       ctx.MustGet(m.NamespaceKey).(string),
		"Secret":          ctx.Param(secretParam),
	}).Debug("Get secret Call")

	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	secret, err := kube.GetSecret(ctx.MustGet(m.NamespaceKey).(string), ctx.Param(secretParam))
	if err != nil {
		ctx.Error(err)
		model.ParseResourceError(err, cherry.ErrUnableGetResource()).Gonic(ctx)
		return
	}

	ret, err := model.ParseSecret(secret)
	if err != nil {
		ctx.Error(err)
		cherry.ErrUnableGetResource().Gonic(ctx)
		return
	}

	ctx.JSON(http.StatusOK, ret)
}

func CreateSecret(ctx *gin.Context) {
	log.WithFields(log.Fields{
		"Namespace": ctx.Param(namespaceParam),
	}).Debug("Create secret Call")

	kubecli := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	var secret model.SecretWithOwner
	if err := ctx.ShouldBindWith(&secret, binding.JSON); err != nil {
		ctx.Error(err)
		cherry.ErrRequestValidationFailed().Gonic(ctx)
		return
	}

	quota, err := kubecli.GetNamespaceQuota(ctx.Param(namespaceParam))
	if err != nil {
		ctx.Error(err)
		model.ParseResourceError(err, cherry.ErrUnableCreateResource()).Gonic(ctx)
		return
	}

	newSecret, errs := model.MakeSecret(ctx.Param(namespaceParam), secret, quota.Labels)
	if errs != nil {
		cherry.ErrRequestValidationFailed().AddDetailsErr(errs...).Gonic(ctx)
		return
	}

	secretAfter, err := kubecli.CreateSecret(newSecret)
	if err != nil {
		ctx.Error(err)
		model.ParseResourceError(err, cherry.ErrUnableCreateResource()).Gonic(ctx)
		return
	}

	ret, err := model.ParseSecret(secretAfter)
	if err != nil {
		ctx.Error(err)
	}

	ctx.JSON(http.StatusCreated, ret)
}

func UpdateSecret(ctx *gin.Context) {
	log.WithFields(log.Fields{
		"Namespace": ctx.Param(namespaceParam),
		"Secret":    ctx.Param(secretParam),
	}).Debug("Create secret Call")

	kubecli := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)

	var secret model.SecretWithOwner
	if err := ctx.ShouldBindWith(&secret, binding.JSON); err != nil {
		ctx.Error(err)
		cherry.ErrRequestValidationFailed().Gonic(ctx)
		return
	}

	quota, err := kubecli.GetNamespaceQuota(ctx.Param(namespaceParam))
	if err != nil {
		ctx.Error(err)
		model.ParseResourceError(err, cherry.ErrUnableUpdateResource()).Gonic(ctx)
		return
	}

	secret.Name = ctx.Param(secretParam)

	newSecret, errs := model.MakeSecret(ctx.Param(namespaceParam), secret, quota.Labels)
	if errs != nil {
		cherry.ErrRequestValidationFailed().AddDetailsErr(errs...).Gonic(ctx)
		return
	}

	secretAfter, err := kubecli.CreateSecret(newSecret)
	if err != nil {
		ctx.Error(err)
		model.ParseResourceError(err, cherry.ErrUnableUpdateResource()).Gonic(ctx)
		return
	}

	ret, err := model.ParseSecret(secretAfter)
	if err != nil {
		ctx.Error(err)
	}

	ctx.JSON(http.StatusAccepted, ret)
}

func DeleteSecret(ctx *gin.Context) {
	log.WithFields(log.Fields{
		"Namespace": ctx.Param(namespaceParam),
		"Secret":    ctx.Param(secretParam),
	}).Debug("Delete secret Call")
	kube := ctx.MustGet(m.KubeClient).(*kubernetes.Kube)
	err := kube.DeleteSecret(ctx.Param(namespaceParam), ctx.Param(secretParam))
	if err != nil {
		ctx.Error(err)
		model.ParseResourceError(err, cherry.ErrUnableDeleteResource()).Gonic(ctx)
		return
	}
	ctx.Status(http.StatusAccepted)
}
