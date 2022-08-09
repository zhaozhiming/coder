import { FC } from "react"
import ReactDiffViewer, { DiffMethod } from "react-diff-viewer"
import { colors } from "theme/colors"

const oldCode = `
workspace_name: "alice-workspace"
workspace_auto_off: true,
template_version_id: "62dee21f-c66785765a8753a9f2aa6786"
`
const newCode = `
workspace_name: "aharvey"
workspace_auto_off: false,
template_version_id: "6287b30c-b9bd666d3ec9d0a9e067c27f"
`

type ExtendedDiffViewer = React.Component

const Diffed = ReactDiffViewer as any as {
  new (): ExtendedDiffViewer
}

const newStyles = {
  variables: {
    dark: {
      removedBackground: colors.red[16],
      removedGutterBackground: colors.red[16],
      addedBackground: colors.green[16],
      addedGutterBackground: colors.green[16],
      // wordAddedBackground: colors.green[16],
    },
  },
  line: {
    padding: "10px 2px",
    "&:hover": {
      background: "#a26ea1",
    },
  },
}

const props: any = {
  oldValue: oldCode,
  newValue: newCode,
  splitView: true,
  useDarkTheme: true,
  styles: newStyles,
  compareMethod: DiffMethod.WORDS,
}

export const Test: FC = () => {
  return <Diffed {...props} />
}
