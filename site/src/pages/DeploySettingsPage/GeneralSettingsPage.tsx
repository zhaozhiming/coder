import Table from "@material-ui/core/Table"
import TableBody from "@material-ui/core/TableBody"
import TableCell from "@material-ui/core/TableCell"
import TableContainer from "@material-ui/core/TableContainer"
import TableHead from "@material-ui/core/TableHead"
import TableRow from "@material-ui/core/TableRow"
import {
  DeploySettingsLayout,
  SettingsHeader,
} from "components/DeploySettingsLayout/DeploySettingsLayout"
import {
  OptionDescription,
  OptionName,
  OptionValue,
} from "components/DeploySettingsLayout/Option"
import React from "react"

export const GeneralSettingsPage: React.FC = () => {
  return (
    <DeploySettingsLayout>
      <SettingsHeader
        title="General"
        description="Deployment and networking settings"
        docsHref="https://coder.com/docs/coder-oss/latest/admin/auth#openid-connect-with-google"
      />

      <TableContainer>
        <Table>
          <TableHead>
            <TableRow>
              <TableCell width="50%">Option</TableCell>
              <TableCell width="50%">Value</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            <TableRow>
              <TableCell>
                <OptionName>Access URL</OptionName>
                <OptionDescription>
                  The address to serve the API and dashboard.
                </OptionDescription>
              </TableCell>

              <TableCell>
                <OptionValue>127.0.0.1:3000</OptionValue>
              </TableCell>
            </TableRow>
            <TableRow>
              <TableCell>
                <OptionName>Wildcard Access URL</OptionName>
                <OptionDescription>
                  Specifies the external URL to access Coder.
                </OptionDescription>
              </TableCell>

              <TableCell>
                <OptionValue>https://www.dev.coder.com</OptionValue>
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </TableContainer>
    </DeploySettingsLayout>
  )
}
