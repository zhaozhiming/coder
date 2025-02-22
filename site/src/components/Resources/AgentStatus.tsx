import Tooltip from "@material-ui/core/Tooltip"
import { makeStyles } from "@material-ui/core/styles"
import { combineClasses } from "util/combineClasses"
import { Workspace, WorkspaceAgent } from "api/typesGenerated"
import { ChooseOne, Cond } from "components/Conditionals/ChooseOne"
import { useTranslation } from "react-i18next"
import WarningRounded from "@material-ui/icons/WarningRounded"
import {
  HelpPopover,
  HelpTooltipText,
  HelpTooltipTitle,
} from "components/Tooltips/HelpTooltip"
import { useRef, useState } from "react"
import Link from "@material-ui/core/Link"

const ConnectedStatus: React.FC = () => {
  const styles = useStyles()
  const { t } = useTranslation("workspacePage")

  return (
    <div
      role="status"
      aria-label={t("agentStatus.connected")}
      className={combineClasses([styles.status, styles.connected])}
    />
  )
}

const DisconnectedStatus: React.FC = () => {
  const styles = useStyles()
  const { t } = useTranslation("workspacePage")

  return (
    <Tooltip title={t("agentStatus.disconnected")}>
      <div
        role="status"
        aria-label={t("agentStatus.disconnected")}
        className={combineClasses([styles.status, styles.disconnected])}
      />
    </Tooltip>
  )
}

const ConnectingStatus: React.FC = () => {
  const styles = useStyles()
  const { t } = useTranslation("workspacePage")

  return (
    <Tooltip title={t("agentStatus.connecting")}>
      <div
        role="status"
        aria-label={t("agentStatus.connecting")}
        className={combineClasses([styles.status, styles.connecting])}
      />
    </Tooltip>
  )
}

const TimeoutStatus: React.FC<{
  agent: WorkspaceAgent
  workspace: Workspace
}> = ({ agent, workspace }) => {
  const { t } = useTranslation("agent")
  const styles = useStyles()
  const anchorRef = useRef<SVGSVGElement>(null)
  const [isOpen, setIsOpen] = useState(false)
  const id = isOpen ? "timeout-popover" : undefined
  const troubleshootLink =
    agent.troubleshooting_url ?? `/templates/${workspace.template_name}#readme`

  return (
    <>
      <WarningRounded
        ref={anchorRef}
        onMouseEnter={() => setIsOpen(true)}
        onMouseLeave={() => setIsOpen(false)}
        role="status"
        aria-label={t("status.timeout")}
        className={styles.timeoutWarning}
      />
      <HelpPopover
        id={id}
        open={isOpen}
        anchorEl={anchorRef.current}
        onOpen={() => setIsOpen(true)}
        onClose={() => setIsOpen(false)}
      >
        <HelpTooltipTitle>{t("timeoutTooltip.title")}</HelpTooltipTitle>
        <HelpTooltipText>
          {t("timeoutTooltip.message")}{" "}
          <Link target="_blank" rel="noreferrer" href={troubleshootLink}>
            {t("timeoutTooltip.link")}
          </Link>
          .
        </HelpTooltipText>
      </HelpPopover>
    </>
  )
}

export const AgentStatus: React.FC<{
  agent: WorkspaceAgent
  workspace: Workspace
}> = ({ agent, workspace }) => {
  return (
    <ChooseOne>
      <Cond condition={agent.status === "connected"}>
        <ConnectedStatus />
      </Cond>
      <Cond condition={agent.status === "disconnected"}>
        <DisconnectedStatus />
      </Cond>
      <Cond condition={agent.status === "timeout"}>
        <TimeoutStatus agent={agent} workspace={workspace} />
      </Cond>
      <Cond>
        <ConnectingStatus />
      </Cond>
    </ChooseOne>
  )
}

const useStyles = makeStyles((theme) => ({
  status: {
    width: theme.spacing(1),
    height: theme.spacing(1),
    borderRadius: "100%",
  },

  connected: {
    backgroundColor: theme.palette.success.light,
    boxShadow: `0 0 12px 0 ${theme.palette.success.light}`,
  },

  disconnected: {
    backgroundColor: theme.palette.text.secondary,
  },

  "@keyframes pulse": {
    "0%": {
      opacity: 1,
    },
    "50%": {
      opacity: 0.4,
    },
    "100%": {
      opacity: 1,
    },
  },

  connecting: {
    backgroundColor: theme.palette.info.light,
    animation: "$pulse 1.5s 0.5s ease-in-out forwards infinite",
  },

  timeoutWarning: {
    color: theme.palette.warning.light,
    width: theme.spacing(2.5),
    height: theme.spacing(2.5),
    position: "relative",
    top: theme.spacing(1),
  },
}))
