/*
Copyright 2021 The Kubermatic Kubernetes Platform contributors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package clustertemplate

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-kit/kit/endpoint"
	"github.com/gorilla/mux"
	"gopkg.in/yaml.v3"

	apiv1 "k8c.io/dashboard/v2/pkg/api/v1"
	apiv2 "k8c.io/dashboard/v2/pkg/api/v2"
	handlercommon "k8c.io/dashboard/v2/pkg/handler/common"
	"k8c.io/dashboard/v2/pkg/handler/middleware"
	"k8c.io/dashboard/v2/pkg/handler/v1/common"
	"k8c.io/dashboard/v2/pkg/handler/v1/label"
	clusterv2 "k8c.io/dashboard/v2/pkg/handler/v2/cluster"
	"k8c.io/dashboard/v2/pkg/provider"
	kubernetesprovider "k8c.io/dashboard/v2/pkg/provider/kubernetes"
	kubermaticv1 "k8c.io/kubermatic/sdk/v2/apis/kubermatic/v1"
	"k8c.io/kubermatic/v2/pkg/features"
	kuberneteshelper "k8c.io/kubermatic/v2/pkg/kubernetes"
	utilerrors "k8c.io/kubermatic/v2/pkg/util/errors"
	"k8c.io/kubermatic/v2/pkg/version"
	clusterv1alpha1 "k8c.io/machine-controller/sdk/apis/cluster/v1alpha1"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
)

// scopeList holds a list of user cluster template access levels.
var scopeList = []string{
	kubermaticv1.UserClusterTemplateScope,
	kubermaticv1.ProjectClusterTemplateScope,
	kubermaticv1.GlobalClusterTemplateScope,
}

const (
	yamlFormat = "yaml"
	jsonFormat = "json"
)

func CreateEndpoint(
	projectProvider provider.ProjectProvider,
	privilegedProjectProvider provider.PrivilegedProjectProvider,
	userInfoGetter provider.UserInfoGetter,
	clusterTemplateProvider provider.ClusterTemplateProvider,
	seedsGetter provider.SeedsGetter,
	credentialManager provider.PresetProvider,
	caBundle *x509.CertPool,
	exposeStrategy kubermaticv1.ExposeStrategy,
	sshKeyProvider provider.SSHKeyProvider,
	configGetter provider.KubermaticConfigurationGetter,
	features features.FeatureGate,
	settingsProvider provider.SettingsProvider,
) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(createClusterTemplateReq)

		config, err := configGetter(ctx)
		if err != nil {
			return nil, err
		}

		err = req.Validate(version.NewFromConfiguration(config))
		if err != nil {
			return nil, apierrors.NewBadRequest(err.Error())
		}

		return createOrUpdateClusterTemplate(ctx, userInfoGetter, seedsGetter, projectProvider, privilegedProjectProvider, sshKeyProvider, credentialManager, exposeStrategy, caBundle, configGetter, features, clusterTemplateProvider, req.Body.CreateClusterSpec, req.ProjectID, req.Body.Name, req.Body.Scope, req.Body.UserSSHKeys, "", settingsProvider)
	}
}

// Validate validates addReq request.
func (req createClusterTemplateReq) Validate(updateManager common.UpdateManager) error {
	if len(req.ProjectID) == 0 || len(req.Body.Name) == 0 || len(req.Body.Scope) == 0 {
		return fmt.Errorf("the name, project ID and scope cannot be empty")
	}

	if err := handlercommon.ValidateClusterSpec(updateManager, req.Body.CreateClusterSpec); err != nil {
		return err
	}

	for _, scope := range scopeList {
		if scope == req.Body.Scope {
			return nil
		}
	}
	return fmt.Errorf("invalid scope name %s", req.Body.Scope)
}

// createClusterTemplateReq defines HTTP request for createClusterTemplate
// swagger:parameters createClusterTemplate
type createClusterTemplateReq struct {
	common.ProjectReq
	// in: body
	Body struct {
		Name        string                        `json:"name"`
		Scope       string                        `json:"scope"`
		UserSSHKeys []apiv2.ClusterTemplateSSHKey `json:"userSshKeys"`
		apiv1.CreateClusterSpec
	}

	// private field for the seed name. Needed for the cluster provider.
	seedName string
}

// GetSeedCluster returns the SeedCluster object.
func (req createClusterTemplateReq) GetSeedCluster() apiv1.SeedCluster {
	return apiv1.SeedCluster{
		SeedName: req.seedName,
	}
}

func DecodeCreateReq(c context.Context, r *http.Request) (interface{}, error) {
	var req createClusterTemplateReq

	pr, err := common.DecodeProjectRequest(c, r)
	if err != nil {
		return nil, err
	}
	req.ProjectReq = pr.(common.ProjectReq)

	if err := json.NewDecoder(r.Body).Decode(&req.Body); err != nil {
		return nil, err
	}

	if len(req.Body.Cluster.Type) == 0 {
		req.Body.Cluster.Type = apiv1.KubernetesClusterType
	}

	seedName, err := clusterv2.FindSeedNameForDatacenter(c, req.Body.Cluster.Spec.Cloud.DatacenterName)
	if err != nil {
		return nil, err
	}
	req.seedName = seedName

	return req, nil
}

// updateClusterTemplateReq defines HTTP request for updateClusterTemplate
// swagger:parameters updateClusterTemplate
type updateClusterTemplateReq struct {
	createClusterTemplateReq

	// in: path
	// required: true
	ClusterTemplateID string `json:"template_id"`
}

func DecodeUpdateReq(c context.Context, r *http.Request) (interface{}, error) {
	var req updateClusterTemplateReq

	pr, err := common.DecodeProjectRequest(c, r)
	if err != nil {
		return nil, err
	}
	req.ProjectReq = pr.(common.ProjectReq)

	if err := json.NewDecoder(r.Body).Decode(&req.Body); err != nil {
		return nil, err
	}

	if len(req.Body.Cluster.Type) == 0 {
		req.Body.Cluster.Type = apiv1.KubernetesClusterType
	}

	seedName, err := clusterv2.FindSeedNameForDatacenter(c, req.Body.Cluster.Spec.Cloud.DatacenterName)
	if err != nil {
		return nil, err
	}
	req.seedName = seedName
	req.ClusterTemplateID = mux.Vars(r)["template_id"]

	return req, nil
}

// GetSeedCluster returns the SeedCluster object.
func (req updateClusterTemplateReq) GetSeedCluster() apiv1.SeedCluster {
	return apiv1.SeedCluster{
		SeedName: req.seedName,
	}
}

// Validate validates updateClusterTemplateReq request.
func (req updateClusterTemplateReq) Validate(updateManager common.UpdateManager) error {
	if len(req.ProjectID) == 0 || len(req.Body.Name) == 0 || len(req.Body.Scope) == 0 {
		return fmt.Errorf("the name, project ID and scope cannot be empty")
	}

	if err := handlercommon.ValidateClusterSpec(updateManager, req.Body.CreateClusterSpec); err != nil {
		return err
	}

	for _, scope := range scopeList {
		if scope == req.Body.Scope {
			return nil
		}
	}
	return fmt.Errorf("invalid scope name %s", req.Body.Scope)
}

func UpdateEndpoint(
	projectProvider provider.ProjectProvider,
	privilegedProjectProvider provider.PrivilegedProjectProvider,
	userInfoGetter provider.UserInfoGetter,
	clusterTemplateProvider provider.ClusterTemplateProvider,
	seedsGetter provider.SeedsGetter,
	credentialManager provider.PresetProvider,
	caBundle *x509.CertPool,
	exposeStrategy kubermaticv1.ExposeStrategy,
	sshKeyProvider provider.SSHKeyProvider,
	configGetter provider.KubermaticConfigurationGetter,
	features features.FeatureGate,
	settingsProvider provider.SettingsProvider,
) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(updateClusterTemplateReq)

		config, err := configGetter(ctx)
		if err != nil {
			return nil, err
		}

		err = req.Validate(version.NewFromConfiguration(config))
		if err != nil {
			return nil, apierrors.NewBadRequest(err.Error())
		}

		return createOrUpdateClusterTemplate(ctx, userInfoGetter, seedsGetter, projectProvider, privilegedProjectProvider, sshKeyProvider, credentialManager, exposeStrategy, caBundle, configGetter, features, clusterTemplateProvider, req.Body.CreateClusterSpec, req.ProjectID, req.Body.Name, req.Body.Scope, req.Body.UserSSHKeys, req.ClusterTemplateID, settingsProvider)
	}
}

func ListEndpoint(projectProvider provider.ProjectProvider, privilegedProjectProvider provider.PrivilegedProjectProvider,
	userInfoGetter provider.UserInfoGetter, clusterTemplateProvider provider.ClusterTemplateProvider) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(listClusterTemplatesReq)
		if err := req.Validate(); err != nil {
			return nil, apierrors.NewBadRequest(err.Error())
		}
		project, err := common.GetProject(ctx, userInfoGetter, projectProvider, privilegedProjectProvider, req.ProjectID, &provider.ProjectGetOptions{IncludeUninitialized: false})
		if err != nil {
			return nil, common.KubernetesErrorToHTTPError(err)
		}
		userInfo, err := userInfoGetter(ctx, "")
		if err != nil {
			return nil, common.KubernetesErrorToHTTPError(err)
		}

		result := apiv2.ClusterTemplateList{}

		templates, err := clusterTemplateProvider.List(ctx, userInfo, project.Name)
		if err != nil {
			return nil, common.KubernetesErrorToHTTPError(err)
		}

		var errorList []string
		for _, template := range templates {
			externalClusterTemplate, err := convertInternalClusterTemplatetoExternal(&template)
			if err != nil {
				errorList = append(errorList, err.Error())
				continue
			}
			result = append(result, *externalClusterTemplate)
		}
		if len(errorList) > 0 {
			return nil, utilerrors.NewWithDetails(http.StatusInternalServerError, "failed to get some cluster templates, please examine details field for more info", errorList)
		}
		return result, nil
	}
}

// listClusterTemplateReq defines HTTP request for listClusterTemplates
// swagger:parameters listClusterTemplates
type listClusterTemplatesReq struct {
	common.ProjectReq
}

func DecodeListReq(c context.Context, r *http.Request) (interface{}, error) {
	var req listClusterTemplatesReq

	pr, err := common.DecodeProjectRequest(c, r)
	if err != nil {
		return nil, err
	}
	req.ProjectReq = pr.(common.ProjectReq)

	return req, nil
}

// Validate validates listClusterTemplatesReq request.
func (req listClusterTemplatesReq) Validate() error {
	if len(req.ProjectID) == 0 {
		return fmt.Errorf("project ID cannot be empty")
	}
	return nil
}

func GetEndpoint(projectProvider provider.ProjectProvider, privilegedProjectProvider provider.PrivilegedProjectProvider,
	userInfoGetter provider.UserInfoGetter, clusterTemplateProvider provider.ClusterTemplateProvider) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(getClusterTemplatesReq)
		if err := req.Validate(); err != nil {
			return nil, apierrors.NewBadRequest(err.Error())
		}

		return getClusterTemplate(ctx, projectProvider, privilegedProjectProvider, userInfoGetter, clusterTemplateProvider, req.ProjectID, req.ClusterTemplateID)
	}
}

func getClusterTemplate(ctx context.Context, projectProvider provider.ProjectProvider, privilegedProjectProvider provider.PrivilegedProjectProvider,
	userInfoGetter provider.UserInfoGetter, clusterTemplateProvider provider.ClusterTemplateProvider, projectID, clusterTemplateID string) (*apiv2.ClusterTemplate, error) {
	project, err := common.GetProject(ctx, userInfoGetter, projectProvider, privilegedProjectProvider, projectID, &provider.ProjectGetOptions{IncludeUninitialized: false})
	if err != nil {
		return nil, common.KubernetesErrorToHTTPError(err)
	}

	userInfo, err := userInfoGetter(ctx, projectID)
	if err != nil {
		return nil, common.KubernetesErrorToHTTPError(err)
	}
	template, err := clusterTemplateProvider.Get(ctx, userInfo, project.Name, clusterTemplateID)
	if err != nil {
		return nil, common.KubernetesErrorToHTTPError(err)
	}
	return convertInternalClusterTemplatetoExternal(template)
}

func ExportEndpoint(projectProvider provider.ProjectProvider, privilegedProjectProvider provider.PrivilegedProjectProvider,
	userInfoGetter provider.UserInfoGetter, clusterTemplateProvider provider.ClusterTemplateProvider) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(exportClusterTemplatesReq)
		if err := req.Validate(); err != nil {
			return nil, apierrors.NewBadRequest(err.Error())
		}

		clusterTemplate, err := getClusterTemplate(ctx, projectProvider, privilegedProjectProvider, userInfoGetter, clusterTemplateProvider, req.ProjectID, req.ClusterTemplateID)
		if err != nil {
			return nil, err
		}

		clusterTemplate.ID = ""
		clusterTemplate.ProjectID = ""
		if clusterTemplate.Cluster.Labels != nil {
			delete(clusterTemplate.Cluster.Labels, kubermaticv1.ProjectIDLabelKey)
		}
		clusterTemplate.Cluster.Credential = ""

		return &encodeClusterTemplateResponse{
			clusterTemplate: clusterTemplate,
			fileSuffix:      req.ClusterTemplateID,
			format:          req.Format,
		}, nil
	}
}

func ImportEndpoint(
	projectProvider provider.ProjectProvider,
	privilegedProjectProvider provider.PrivilegedProjectProvider,
	userInfoGetter provider.UserInfoGetter,
	clusterTemplateProvider provider.ClusterTemplateProvider,
	seedsGetter provider.SeedsGetter,
	credentialManager provider.PresetProvider,
	caBundle *x509.CertPool,
	exposeStrategy kubermaticv1.ExposeStrategy,
	sshKeyProvider provider.SSHKeyProvider,
	configGetter provider.KubermaticConfigurationGetter,
	features features.FeatureGate,
	settingsProvider provider.SettingsProvider,
) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(importClusterTemplateReq)

		config, err := configGetter(ctx)
		if err != nil {
			return nil, err
		}

		err = req.Validate(version.NewFromConfiguration(config))
		if err != nil {
			return nil, apierrors.NewBadRequest(err.Error())
		}

		var nd *apiv1.NodeDeployment
		if req.Body.NodeDeployment != nil {
			nd = &apiv1.NodeDeployment{
				Spec: req.Body.NodeDeployment.Spec,
			}
		}

		apps := req.getApplicationsFromRequest()

		createCluster := apiv1.CreateClusterSpec{
			Cluster: apiv1.Cluster{
				ObjectMeta: apiv1.ObjectMeta{
					Name: req.Body.Name,
				},
				Labels:          req.Body.Cluster.Labels,
				InheritedLabels: req.Body.Cluster.InheritedLabels,
				Type:            apiv1.KubernetesClusterType,
				Credential:      req.Body.Cluster.Credential,
				Spec:            req.Body.Cluster.Spec,
			},
			NodeDeployment: nd,
			Applications:   apps,
		}

		return createOrUpdateClusterTemplate(ctx, userInfoGetter, seedsGetter, projectProvider, privilegedProjectProvider, sshKeyProvider, credentialManager, exposeStrategy, caBundle, configGetter, features, clusterTemplateProvider, createCluster, req.ProjectID, req.Body.Name, req.Body.Scope, req.Body.UserSSHKeys, "", settingsProvider)
	}
}

func createOrUpdateClusterTemplate(ctx context.Context, userInfoGetter provider.UserInfoGetter, seedsGetter provider.SeedsGetter, projectProvider provider.ProjectProvider, privilegedProjectProvider provider.PrivilegedProjectProvider, sshKeyProvider provider.SSHKeyProvider, credentialManager provider.PresetProvider, exposeStrategy kubermaticv1.ExposeStrategy, caBundle *x509.CertPool, configGetter provider.KubermaticConfigurationGetter, features features.FeatureGate, clusterTemplateProvider provider.ClusterTemplateProvider, createCluster apiv1.CreateClusterSpec, projectID, name, scope string, userSSHKeys []apiv2.ClusterTemplateSSHKey, clusterTemplateID string, settingsProvider provider.SettingsProvider) (*apiv2.ClusterTemplate, error) {
	privilegedClusterProvider := ctx.Value(middleware.PrivilegedClusterProviderContextKey).(provider.PrivilegedClusterProvider)

	project, err := common.GetProject(ctx, userInfoGetter, projectProvider, privilegedProjectProvider, projectID, &provider.ProjectGetOptions{IncludeUninitialized: false})
	if err != nil {
		return nil, common.KubernetesErrorToHTTPError(err)
	}
	userInfo, err := userInfoGetter(ctx, projectID)
	if err != nil {
		return nil, common.KubernetesErrorToHTTPError(err)
	}

	partialCluster, err := handlercommon.GenerateCluster(ctx, projectID, createCluster, seedsGetter, credentialManager, exposeStrategy, userInfoGetter, caBundle, configGetter, features, settingsProvider)
	if err != nil {
		return nil, err
	}

	isUpdateRequest := false
	if clusterTemplateID != "" {
		isUpdateRequest = true
		partialCluster.Name = clusterTemplateID
	}

	newClusterTemplate := &kubermaticv1.ClusterTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Name:        partialCluster.Name,
			Labels:      map[string]string{},
			Annotations: map[string]string{},
		},
		Credential:             partialCluster.GetSecretName(),
		ClusterLabels:          partialCluster.Labels,
		InheritedClusterLabels: createCluster.Cluster.InheritedLabels,
		Spec:                   partialCluster.Spec,
	}

	if err := kubernetesprovider.CreateOrUpdateCredentialSecretForCluster(ctx, privilegedClusterProvider.GetSeedClusterAdminRuntimeClient(), partialCluster); err != nil {
		return nil, err
	}

	isBYO, err := common.IsBringYourOwnProvider(partialCluster.Spec.Cloud)
	if err != nil {
		return nil, common.KubernetesErrorToHTTPError(err)
	}
	if !isBYO {
		kuberneteshelper.AddFinalizer(newClusterTemplate, kubermaticv1.CredentialsSecretsCleanupFinalizer)
	}

	// copy preset annotations
	if partialCluster.Annotations != nil {
		newClusterTemplate.Annotations = partialCluster.Annotations
	}

	newClusterTemplate.Annotations[kubermaticv1.InitialMachineDeploymentRequestAnnotation] = partialCluster.Annotations[kubermaticv1.InitialMachineDeploymentRequestAnnotation]

	newClusterTemplate.Annotations[kubermaticv1.ClusterTemplateUserAnnotationKey] = userInfo.Email
	newClusterTemplate.Labels[kubermaticv1.ClusterTemplateProjectLabelKey] = project.Name
	newClusterTemplate.Labels[kubermaticv1.ClusterTemplateScopeLabelKey] = scope
	newClusterTemplate.Labels[kubermaticv1.ClusterTemplateHumanReadableNameLabelKey] = name
	if val, ok := partialCluster.Labels[kubermaticv1.IsCredentialPresetLabelKey]; ok {
		newClusterTemplate.Labels[kubermaticv1.IsCredentialPresetLabelKey] = val
		newClusterTemplate.Annotations[kubermaticv1.PresetNameAnnotation] = partialCluster.Annotations[kubermaticv1.PresetNameAnnotation]
	}

	// SSH check
	if len(userSSHKeys) > 0 && scope == kubermaticv1.ProjectClusterTemplateScope {
		projectSSHKeys, err := sshKeyProvider.List(ctx, project, nil)
		if err != nil {
			return nil, common.KubernetesErrorToHTTPError(err)
		}
		for _, templateKey := range userSSHKeys {
			found := false
			for _, projectSSHKey := range projectSSHKeys {
				if projectSSHKey.Name == templateKey.ID {
					found = true
					break
				}
			}
			if !found {
				return nil, fmt.Errorf("the given ssh key %s does not belong to the given project %s (%s)", templateKey.ID, project.Spec.Name, project.Name)
			}
			newClusterTemplate.UserSSHKeys = append(newClusterTemplate.UserSSHKeys, kubermaticv1.ClusterTemplateSSHKey{
				Name: templateKey.Name,
				ID:   templateKey.ID,
			})
		}
	}

	ct, err := clusterTemplateProvider.CreateorUpdate(ctx, userInfo, newClusterTemplate, scope, project.Name, isUpdateRequest)
	if err != nil {
		return nil, common.KubernetesErrorToHTTPError(err)
	}

	return convertInternalClusterTemplatetoExternal(ct)
}

func DeleteEndpoint(projectProvider provider.ProjectProvider, privilegedProjectProvider provider.PrivilegedProjectProvider,
	userInfoGetter provider.UserInfoGetter, clusterTemplateProvider provider.ClusterTemplateProvider) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(getClusterTemplatesReq)
		if err := req.Validate(); err != nil {
			return nil, apierrors.NewBadRequest(err.Error())
		}
		project, err := common.GetProject(ctx, userInfoGetter, projectProvider, privilegedProjectProvider, req.ProjectID, &provider.ProjectGetOptions{IncludeUninitialized: false})
		if err != nil {
			return nil, common.KubernetesErrorToHTTPError(err)
		}

		userInfo, err := userInfoGetter(ctx, req.ProjectID)
		if err != nil {
			return nil, common.KubernetesErrorToHTTPError(err)
		}
		if err := clusterTemplateProvider.Delete(ctx, userInfo, project.Name, req.ClusterTemplateID); err != nil {
			return nil, common.KubernetesErrorToHTTPError(err)
		}
		return nil, nil
	}
}

func CreateInstanceEndpoint(projectProvider provider.ProjectProvider, privilegedProjectProvider provider.PrivilegedProjectProvider,
	userInfoGetter provider.UserInfoGetter, clusterTemplateProvider provider.ClusterTemplateProvider, seedsGetter provider.SeedsGetter, clusterTemplateProviderGetter provider.ClusterTemplateInstanceProviderGetter) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(createInstanceReq)
		if err := req.Validate(); err != nil {
			return nil, apierrors.NewBadRequest(err.Error())
		}
		project, err := common.GetProject(ctx, userInfoGetter, projectProvider, privilegedProjectProvider, req.ProjectID, &provider.ProjectGetOptions{IncludeUninitialized: false})
		if err != nil {
			return nil, common.KubernetesErrorToHTTPError(err)
		}

		adminUserInfo, err := userInfoGetter(ctx, "")
		if err != nil {
			return nil, common.KubernetesErrorToHTTPError(err)
		}
		ct, err := clusterTemplateProvider.Get(ctx, adminUserInfo, project.Name, req.ClusterTemplateID)
		if err != nil {
			return nil, common.KubernetesErrorToHTTPError(err)
		}

		seed, _, err := provider.DatacenterFromSeedMap(adminUserInfo, seedsGetter, ct.Spec.Cloud.DatacenterName)
		if err != nil {
			return nil, fmt.Errorf("error getting seed: %w", err)
		}

		clusterTemplateInstanceProvider, err := clusterTemplateProviderGetter(seed)
		if err != nil {
			return nil, common.KubernetesErrorToHTTPError(err)
		}

		if adminUserInfo.IsAdmin {
			privilegedclusterTemplateInstanceProvider := clusterTemplateInstanceProvider.(provider.PrivilegedClusterTemplateInstanceProvider)
			instance, err := privilegedclusterTemplateInstanceProvider.CreateUnsecured(ctx, adminUserInfo, ct, project, req.Body.Replicas)
			if err != nil {
				return nil, common.KubernetesErrorToHTTPError(err)
			}
			return apiv2.ClusterTemplateInstance{
				Name: instance.Name,
				Spec: instance.Spec,
			}, nil
		}

		userInfo, err := userInfoGetter(ctx, project.Name)
		if err != nil {
			return nil, common.KubernetesErrorToHTTPError(err)
		}

		instance, err := clusterTemplateInstanceProvider.Create(ctx, userInfo, ct, project, req.Body.Replicas)
		if err != nil {
			return nil, common.KubernetesErrorToHTTPError(err)
		}

		return apiv2.ClusterTemplateInstance{
			Name: instance.Name,
			Spec: instance.Spec,
		}, nil
	}
}

type encodeClusterTemplateResponse struct {
	clusterTemplate *apiv2.ClusterTemplate
	fileSuffix      string
	format          string
}

func EncodeClusterTemplate(_ context.Context, w http.ResponseWriter, response interface{}) (err error) {
	rsp := response.(*encodeClusterTemplateResponse)
	clusterTemplate := rsp.clusterTemplate
	filename := "clusterTemplate"

	if len(rsp.fileSuffix) > 0 {
		filename = fmt.Sprintf("%s-%s", filename, rsp.fileSuffix)
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-disposition", fmt.Sprintf("attachment; filename=%s", filename))
	w.Header().Add("Cache-Control", "no-cache")

	if rsp.format == yamlFormat {
		b, err := yaml.Marshal(clusterTemplate)
		if err != nil {
			return err
		}
		_, err = w.Write(b)
		return err
	}

	b, err := json.Marshal(clusterTemplate)
	if err != nil {
		return err
	}

	_, err = w.Write(b)
	return err
}

// createInstanceReq defines HTTP request for createClusterTemplateInstance
// swagger:parameters createClusterTemplateInstance
type createInstanceReq struct {
	getClusterTemplatesReq
	// in: body
	Body struct {
		Replicas int64 `json:"replicas"`
	}
}

func DecodeCreateInstanceReq(c context.Context, r *http.Request) (interface{}, error) {
	var req createInstanceReq

	pr, err := DecodeGetReq(c, r)
	if err != nil {
		return nil, err
	}
	req.getClusterTemplatesReq = pr.(getClusterTemplatesReq)
	if err := json.NewDecoder(r.Body).Decode(&req.Body); err != nil {
		return nil, err
	}

	return req, nil
}

// getClusterTemplatesReq defines HTTP request for getClusterTemplate
// swagger:parameters getClusterTemplate deleteClusterTemplate
type getClusterTemplatesReq struct {
	common.ProjectReq
	// in: path
	// required: true
	ClusterTemplateID string `json:"template_id"`
}

// Validate validates getClusterTemplatesReq request.
func (req getClusterTemplatesReq) Validate() error {
	if len(req.ProjectID) == 0 {
		return fmt.Errorf("project ID cannot be empty")
	}
	if len(req.ClusterTemplateID) == 0 {
		return fmt.Errorf("cluster template ID cannot be empty")
	}
	return nil
}

func DecodeGetReq(c context.Context, r *http.Request) (interface{}, error) {
	var req getClusterTemplatesReq

	pr, err := common.DecodeProjectRequest(c, r)
	if err != nil {
		return nil, err
	}
	req.ProjectReq = pr.(common.ProjectReq)

	req.ClusterTemplateID = mux.Vars(r)["template_id"]

	return req, nil
}

// exportClusterTemplatesReq defines HTTP request for exportClusterTemplate
// swagger:parameters exportClusterTemplate
type exportClusterTemplatesReq struct {
	getClusterTemplatesReq

	// in: query
	Format string `json:"format,omitempty"`
}

func DecodeExportReq(c context.Context, r *http.Request) (interface{}, error) {
	var req exportClusterTemplatesReq

	getReq, err := DecodeGetReq(c, r)
	if err != nil {
		return nil, err
	}
	req.getClusterTemplatesReq = getReq.(getClusterTemplatesReq)

	queryParam := r.URL.Query().Get("format")

	supportedFormats := sets.New(yamlFormat, jsonFormat)

	if len(queryParam) > 0 && !supportedFormats.Has(queryParam) {
		return nil, fmt.Errorf("not supported file format: %s", queryParam)
	}
	req.Format = queryParam

	return req, nil
}

func convertInternalClusterTemplatetoExternal(template *kubermaticv1.ClusterTemplate) (*apiv2.ClusterTemplate, error) {
	initialNodeDeployment := &apiv1.NodeDeployment{}
	if rawMachineDeployment := template.Annotations[kubermaticv1.InitialMachineDeploymentRequestAnnotation]; rawMachineDeployment != "" {
		var md clusterv1alpha1.MachineDeployment
		err := json.Unmarshal([]byte(rawMachineDeployment), &md)
		if err != nil {
			return nil, err
		}

		// convert MachineDeployment back into a NodeDeployment
		initialNodeDeployment, err = handlercommon.OutputMachineDeployment(&md)
		if err != nil {
			return nil, err
		}
	}

	var apps []apiv1.Application
	rawApplicationsRequest, ok := template.Annotations[kubermaticv1.InitialApplicationInstallationsRequestAnnotation]
	if ok && rawApplicationsRequest != "" {
		err := json.Unmarshal([]byte(rawApplicationsRequest), &apps)
		if err != nil {
			return nil, err
		}
	}

	ct := &apiv2.ClusterTemplate{
		ObjectMeta: apiv1.ObjectMeta{
			ID:                template.Name,
			Name:              template.Spec.HumanReadableName,
			CreationTimestamp: apiv1.NewTime(template.CreationTimestamp.Time),
			Annotations:       template.Annotations,
			DeletionTimestamp: func() *apiv1.Time {
				if template.DeletionTimestamp != nil {
					deletionTimestamp := apiv1.NewTime(template.DeletionTimestamp.Time)
					return &deletionTimestamp
				}
				return nil
			}(),
		},
		Name:      template.Labels[kubermaticv1.ClusterTemplateHumanReadableNameLabelKey],
		ID:        template.Name,
		ProjectID: template.Labels[kubermaticv1.ClusterTemplateProjectLabelKey],
		User:      template.Annotations[kubermaticv1.ClusterTemplateUserAnnotationKey],
		Scope:     template.Labels[kubermaticv1.ClusterTemplateScopeLabelKey],
		Cluster: &apiv2.ClusterTemplateInfo{
			Labels:          label.FilterLabels(label.ClusterResourceType, template.ClusterLabels),
			Annotations:     template.Annotations,
			InheritedLabels: template.InheritedClusterLabels,
			Credential:      template.Credential,
			Spec: apiv1.ClusterSpec{
				Cloud:                                template.Spec.Cloud,
				MachineNetworks:                      template.Spec.MachineNetworks,
				Version:                              template.Spec.Version,
				OIDC:                                 template.Spec.OIDC,
				UpdateWindow:                         template.Spec.UpdateWindow,
				UsePodSecurityPolicyAdmissionPlugin:  template.Spec.UsePodSecurityPolicyAdmissionPlugin,
				UsePodNodeSelectorAdmissionPlugin:    template.Spec.UsePodNodeSelectorAdmissionPlugin,
				UseEventRateLimitAdmissionPlugin:     template.Spec.UseEventRateLimitAdmissionPlugin,
				EnableUserSSHKeyAgent:                template.Spec.EnableUserSSHKeyAgent,
				BackupConfig:                         template.Spec.BackupConfig,
				KubeLB:                               template.Spec.KubeLB,
				KubernetesDashboard:                  template.Spec.KubernetesDashboard,
				PodNodeSelectorAdmissionPluginConfig: template.Spec.PodNodeSelectorAdmissionPluginConfig,
				EventRateLimitConfig:                 template.Spec.EventRateLimitConfig,
				AdmissionPlugins:                     template.Spec.AdmissionPlugins,
				AuditLogging:                         template.Spec.AuditLogging,
				ServiceAccount:                       template.Spec.ServiceAccount,
				OPAIntegration:                       template.Spec.OPAIntegration,
				MLA:                                  template.Spec.MLA,
				ClusterNetwork:                       &template.Spec.ClusterNetwork,
				ContainerRuntime:                     template.Spec.ContainerRuntime,
				CNIPlugin:                            template.Spec.CNIPlugin,
				ExposeStrategy:                       template.Spec.ExposeStrategy,
				APIServerAllowedIPRanges:             template.Spec.APIServerAllowedIPRanges,
				DisableCSIDriver:                     template.Spec.DisableCSIDriver,
				Kyverno:                              template.Spec.Kyverno,
			},
		},
		NodeDeployment: &apiv2.ClusterTemplateNodeDeployment{
			Spec:        initialNodeDeployment.Spec,
			Annotations: initialNodeDeployment.Annotations,
		},
		Applications: apps,
	}

	if len(template.UserSSHKeys) > 0 {
		for _, sshKey := range template.UserSSHKeys {
			ct.UserSSHKeys = append(ct.UserSSHKeys, apiv2.ClusterTemplateSSHKey{
				Name: sshKey.Name,
				ID:   sshKey.ID,
			})
		}
	}

	return ct, nil
}

// importClusterTemplateReq defines HTTP requests for importClusterTemplate
// swagger:parameters importClusterTemplate
type importClusterTemplateReq struct {
	common.ProjectReq
	// in: body
	Body struct {
		apiv2.ClusterTemplate
	}

	// private field for the seed name. Needed for the cluster provider.
	seedName string
}

// GetSeedCluster returns the SeedCluster object.
func (req importClusterTemplateReq) GetSeedCluster() apiv1.SeedCluster {
	return apiv1.SeedCluster{
		SeedName: req.seedName,
	}
}

// Validate validates addReq request.
func (req importClusterTemplateReq) Validate(updateManager common.UpdateManager) error {
	if req.Body.Cluster == nil {
		return fmt.Errorf("the cluster cannot be empty")
	}
	if len(req.ProjectID) == 0 || len(req.Body.Name) == 0 || len(req.Body.Scope) == 0 {
		return fmt.Errorf("the name, project ID and scope cannot be empty")
	}

	var nd *apiv1.NodeDeployment
	if req.Body.NodeDeployment != nil {
		nd = &apiv1.NodeDeployment{
			Spec: req.Body.NodeDeployment.Spec,
		}
	}

	apps := req.getApplicationsFromRequest()

	if err := handlercommon.ValidateClusterSpec(updateManager, apiv1.CreateClusterSpec{
		Cluster: apiv1.Cluster{
			Type: apiv1.KubernetesClusterType,
			Spec: req.Body.Cluster.Spec,
		},
		NodeDeployment: nd,
		Applications:   apps,
	}); err != nil {
		return err
	}

	for _, scope := range scopeList {
		if scope == req.Body.Scope {
			return nil
		}
	}
	return fmt.Errorf("invalid scope name %s", req.Body.Scope)
}

func (req importClusterTemplateReq) getApplicationsFromRequest() []apiv1.Application {
	var applications []apiv1.Application
	for _, app := range req.Body.Applications {
		newApp := apiv1.Application{
			Spec: app.Spec,
		}

		applications = append(applications, newApp)
	}
	return applications
}

func DecodeImportReq(c context.Context, r *http.Request) (interface{}, error) {
	var req importClusterTemplateReq

	pr, err := common.DecodeProjectRequest(c, r)
	if err != nil {
		return nil, err
	}
	req.ProjectReq = pr.(common.ProjectReq)

	if err := json.NewDecoder(r.Body).Decode(&req.Body); err != nil {
		return nil, err
	}

	seedName, err := clusterv2.FindSeedNameForDatacenter(c, req.Body.Cluster.Spec.Cloud.DatacenterName)
	if err != nil {
		return nil, err
	}
	req.seedName = seedName

	return req, nil
}
