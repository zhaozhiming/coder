import { useActor } from "@xstate/react"
import { User, Template, ParameterSchema } from "api/typesGenerated"
import { useOrganizationId } from "hooks/useOrganizationId"
import { FC, useContext, useEffect, useState } from "react"
import { Helmet } from "react-helmet-async"
import { useNavigate, useParams } from "react-router-dom"
import { pageTitle } from "util/page"
import { XServiceContext } from "xServices/StateContext"
import { CreateWorkspaceErrors, CreateWorkspacePageView } from "./CreateWorkspacePageView"
import {
  checkAuthorization,
  createWorkspace,
  getTemplates,
  getTemplateVersionSchema,
} from "api/api"

const CreateWorkspacePage: FC = () => {
  const { template } = useParams()
  const navigate = useNavigate()
  const xServices = useContext(XServiceContext)
  const [authState] = useActor(xServices.authXService)
  const { me } = authState.context

  const [organizationId] = useState<string>(useOrganizationId())
  const [owner, setOwner] = useState<User | null>((me ?? null))
  const [templateName] = useState<string>(template ? template : "")
  const [templates] = useState<Template[]>()
  const [selectedTemplate, setSelectedTemplate] = useState<Template>()
  const [templateSchema, setTemplateSchema] = useState<ParameterSchema[]>()
  const [creatingWorkspace, setCreatingWorkspace] = useState<boolean>(false)
  const [createWorkspaceError, setCreateWorkspaceError] = useState<Error | unknown>()
  const [getTemplatesError, setGetTemplatesError] = useState<Error | unknown>()
  const [getTemplateSchemaError, setGetTemplateSchemaError] = useState<Error | unknown>()
  const [permissions, setPermissions] = useState<Record<string, boolean>>()
  const [checkPermissionsError, setCheckPermissionsError] = useState<Error | unknown>()

  useEffect(() => {
    setGetTemplatesError(undefined)
    getTemplates(organizationId).then((res) => {
      const temps = res.filter((template) => template.name === templateName)
      const selectedTemps = res.length > 0 ? temps[0] : undefined
      setSelectedTemplate(selectedTemps)
    }, (err) => {
      setGetTemplatesError(err)
    })
  }, [organizationId, templateName])

  useEffect(() => {
    if (!selectedTemplate) {
      return
    }

    setGetTemplateSchemaError(undefined)
    getTemplateVersionSchema(selectedTemplate.active_version_id).then((res) => {
      // Only show parameters that are allowed to be overridden.
      // CLI code: https://github.com/coder/coder/blob/main/cli/create.go#L152-L155
      res = res.filter((param) => param.allow_override_source)
      setTemplateSchema(res)
    }, (err) => {
      setGetTemplateSchemaError(err)
    })
  }, [selectedTemplate])

  useEffect(() => {
    if (!organizationId) {
      return
    }

    // HACK: below, we pass in * for the owner_id, which is a hacky way of checking if the
    // current user can create a workspace on behalf of anyone within the org (only org owners should be able to do this).
    // This pattern should not be replicated outside of this narrow use case.
    const permissionsToCheck = {
      createWorkspaceForUser: {
        object: {
          resource_type: "workspace",
          organization_id: `${organizationId}`,
          owner_id: "*",
        },
        action: "create",
      },
    }

    setCheckPermissionsError(undefined)
    checkAuthorization({
      checks: permissionsToCheck,
    }).then((res) => {
      setPermissions(res)
    }, (err) => {
      setCheckPermissionsError(err)
    })
  }, [organizationId, selectedTemplate])

  const hasErrors = getTemplatesError || getTemplateSchemaError || createWorkspaceError ? true : false

  return (
    <>
      <Helmet>
        <title>{pageTitle("Create Workspace")}</title>
      </Helmet>
      <CreateWorkspacePageView
        loadingTemplates={templates === undefined}
        loadingTemplateSchema={templateSchema === undefined}
        creatingWorkspace={creatingWorkspace}
        hasTemplateErrors={hasErrors}
        templateName={templateName}
        templates={templates}
        selectedTemplate={selectedTemplate}
        templateSchema={templateSchema}
        createWorkspaceErrors={{
          [CreateWorkspaceErrors.GET_TEMPLATES_ERROR]: getTemplatesError,
          [CreateWorkspaceErrors.GET_TEMPLATE_SCHEMA_ERROR]: getTemplateSchemaError,
          [CreateWorkspaceErrors.CREATE_WORKSPACE_ERROR]: createWorkspaceError,
          [CreateWorkspaceErrors.CHECK_PERMISSIONS_ERROR]: checkPermissionsError,
        }}
        canCreateForUser={permissions?.createWorkspaceForUser}
        defaultWorkspaceOwner={me ?? null}
        setOwner={setOwner}
        onCancel={() => {
          navigate("/templates")
        }}
        onSubmit={(req) => {
          setCreatingWorkspace(true)

          createWorkspace(organizationId, owner?.id ?? "me", req)
          .then((res) => {
            navigate(`/@${res.owner_name}/${res.name}`)
          }, (err) => {
            setCreateWorkspaceError(err)
          })
        }}
      />
    </>
  )
}

export default CreateWorkspacePage
