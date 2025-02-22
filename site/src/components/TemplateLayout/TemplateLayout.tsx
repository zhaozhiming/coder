import Avatar from "@material-ui/core/Avatar"
import Button from "@material-ui/core/Button"
import Link from "@material-ui/core/Link"
import { makeStyles } from "@material-ui/core/styles"
import AddCircleOutline from "@material-ui/icons/AddCircleOutline"
import SettingsOutlined from "@material-ui/icons/SettingsOutlined"
import { useMachine, useSelector } from "@xstate/react"
import {
  PageHeader,
  PageHeaderSubtitle,
  PageHeaderTitle,
} from "components/PageHeader/PageHeader"
import { useOrganizationId } from "hooks/useOrganizationId"
import {
  createContext,
  FC,
  PropsWithChildren,
  Suspense,
  useContext,
} from "react"
import { Link as RouterLink, NavLink, useParams } from "react-router-dom"
import { combineClasses } from "util/combineClasses"
import { firstLetter } from "util/firstLetter"
import { selectPermissions } from "xServices/auth/authSelectors"
import { XServiceContext } from "xServices/StateContext"
import {
  TemplateContext,
  templateMachine,
} from "xServices/template/templateXService"
import { Margins } from "components/Margins/Margins"
import { Stack } from "components/Stack/Stack"
import { Permissions } from "xServices/auth/authXService"
import { Loader } from "components/Loader/Loader"

const Language = {
  settingsButton: "Settings",
  createButton: "Create workspace",
  noDescription: "",
}

const useTemplateName = () => {
  const { template } = useParams()

  if (!template) {
    throw new Error("No template found in the URL")
  }

  return template
}

type TemplateLayoutContextValue = {
  context: TemplateContext
  permissions?: Permissions
}

const TemplateLayoutContext = createContext<
  TemplateLayoutContextValue | undefined
>(undefined)

export const useTemplateLayoutContext = (): TemplateLayoutContextValue => {
  const context = useContext(TemplateLayoutContext)
  if (!context) {
    throw new Error(
      "useTemplateLayoutContext only can be used inside of TemplateLayout",
    )
  }
  return context
}

const TemplateSettingsButton: FC<{ templateName: string }> = ({
  templateName,
}) => (
  <Link
    underline="none"
    component={RouterLink}
    to={`/templates/${templateName}/settings`}
  >
    <Button variant="outlined" startIcon={<SettingsOutlined />}>
      {Language.settingsButton}
    </Button>
  </Link>
)

const CreateWorkspaceButton: FC<{
  templateName: string
  className?: string
}> = ({ templateName, className }) => (
  <Link
    underline="none"
    component={RouterLink}
    to={`/templates/${templateName}/workspace`}
  >
    <Button className={className ?? ""} startIcon={<AddCircleOutline />}>
      {Language.createButton}
    </Button>
  </Link>
)

export const TemplateLayout: FC<PropsWithChildren> = ({ children }) => {
  const styles = useStyles()
  const organizationId = useOrganizationId()
  const templateName = useTemplateName()
  const [templateState, _] = useMachine(templateMachine, {
    context: {
      templateName,
      organizationId,
    },
  })
  const {
    template,
    activeTemplateVersion,
    templateResources,
    templateDAUs,
    permissions: templatePermissions,
  } = templateState.context
  const xServices = useContext(XServiceContext)
  const permissions = useSelector(xServices.authXService, selectPermissions)
  const isLoading =
    !template ||
    !activeTemplateVersion ||
    !templateResources ||
    !permissions ||
    !templateDAUs ||
    !templatePermissions

  const hasIcon = template && template.icon && template.icon !== ""

  const generatePageHeaderActions = (): JSX.Element[] => {
    const pageActions: JSX.Element[] = []

    if (!isLoading && templatePermissions.canUpdateTemplate) {
      pageActions.push(<TemplateSettingsButton templateName={templateName} />)
    }

    if (!isLoading) {
      pageActions.push(<CreateWorkspaceButton templateName={templateName} />)
    }

    return pageActions
  }

  return (
    <>
      <Margins>
        <PageHeader
          actions={
            <>
              {generatePageHeaderActions().map((action, i) => (
                <div key={i}>{action}</div>
              ))}
            </>
          }
        >
          <Stack direction="row" spacing={3} className={styles.pageTitle}>
            {!isLoading && (
              <div>
                {hasIcon ? (
                  <div className={styles.iconWrapper}>
                    <img src={template.icon} alt="" />
                  </div>
                ) : (
                  <Avatar className={styles.avatar}>
                    {firstLetter(templateName)}
                  </Avatar>
                )}
              </div>
            )}

            {!isLoading && (
              <div>
                <PageHeaderTitle>{templateName}</PageHeaderTitle>
                <PageHeaderSubtitle condensed>
                  {template.description === ""
                    ? Language.noDescription
                    : template.description}
                </PageHeaderSubtitle>
              </div>
            )}
          </Stack>
        </PageHeader>
      </Margins>

      {!isLoading && (
        <div className={styles.tabs}>
          <Margins>
            <Stack direction="row" spacing={0.25}>
              <NavLink
                end
                to={`/templates/${template.name}`}
                className={({ isActive }) =>
                  combineClasses([
                    styles.tabItem,
                    isActive ? styles.tabItemActive : undefined,
                  ])
                }
              >
                Summary
              </NavLink>
              <NavLink
                to={`/templates/${template.name}/permissions`}
                className={({ isActive }) =>
                  combineClasses([
                    styles.tabItem,
                    isActive ? styles.tabItemActive : undefined,
                  ])
                }
              >
                Permissions
              </NavLink>
            </Stack>
          </Margins>
        </div>
      )}

      <Margins>
        <TemplateLayoutContext.Provider
          value={{ permissions, context: templateState.context }}
        >
          <Suspense fallback={<Loader />}>{children}</Suspense>
        </TemplateLayoutContext.Provider>
      </Margins>
    </>
  )
}

export const useStyles = makeStyles((theme) => {
  return {
    pageTitle: {
      alignItems: "center",
    },
    avatar: {
      width: theme.spacing(6),
      height: theme.spacing(6),
      fontSize: theme.spacing(3),
    },
    iconWrapper: {
      width: theme.spacing(6),
      height: theme.spacing(6),
      "& img": {
        width: "100%",
      },
    },

    tabs: {
      borderBottom: `1px solid ${theme.palette.divider}`,
      marginBottom: theme.spacing(5),
    },

    tabItem: {
      textDecoration: "none",
      color: theme.palette.text.secondary,
      fontSize: 14,
      display: "block",
      padding: theme.spacing(0, 2, 2),

      "&:hover": {
        color: theme.palette.text.primary,
      },
    },

    tabItemActive: {
      color: theme.palette.text.primary,
      position: "relative",

      "&:before": {
        content: `""`,
        left: 0,
        bottom: 0,
        height: 2,
        width: "100%",
        background: theme.palette.secondary.dark,
        position: "absolute",
      },
    },
  }
})
