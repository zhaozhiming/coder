# Secrets

<blockquote class="info">
This article explains how to use secrets in a workspace. To authenticate the
workspace provisioner, see <a href="./templates/authentication">this</a>.
</blockquote>

Coder is open-minded about how you get your secrets into your workspaces.

## Wait a minute...

Your first stab at secrets with Coder should be your local method.
You can do everything you can locally and more with your Coder workspace, so
whatever workflow and tools you already use to manage secrets may be brought
over.

Often, this workflow is simply:

1. Give your users their secrets in advance
1. Your users write them to a persistent file after
   they've built their workspace

<a href="./templates#parameters">Template parameters</a> are a dangerous way to accept secrets.
We show parameters in cleartext around the product. Assume anyone with view
access to a workspace can also see its parameters.

## SSH Keys

Coder generates SSH key pairs for each user. This can be used as an authentication mechanism for
git providers or other tools. Within workspaces, git will attempt to use this key within workspaces
via the `$GIT_SSH_COMMAND` environment variable.

Users can view their public key in their account settings:

![SSH keys in account settings](./images/ssh-keys.png)

> There is a [known issue](https://github.com/coder/coder/issues/3126) that prevents users from
> using their own SSH keys within Coder workspaces.

## Dynamic Secrets

Dynamic secrets are attached to the workspace lifecycle and automatically
injected into the workspace. With a little bit of up front template work,
they make life simpler for both the end user and the security team.

This method is limited to
[services with Terraform providers](https://registry.terraform.io/browse/providers),
which excludes obscure API providers.

Dynamic secrets can be implemented in your template code like so:

```hcl
resource "twilio_iam_api_key" "api_key" {
  account_sid   = "ACXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"
  friendly_name = "Test API Key"
}

resource "coder_agent" "main" {
  # ...
  env = {
    # Let users access the secret via $TWILIO_API_SECRET
    TWILIO_API_SECRET = "${twilio_iam_api_key.api_key.secret}"
  }
}
```

A catch-all variation of this approach is dynamically provisioning a cloud service account (e.g [GCP](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/google_service_account_key#private_key))
for each workspace and then making the relevant secrets available via the cloud's secret management
system.

## Token Management (enterprise)

Coder can manage tokens on behalf of users on the following platforms:

- GitHub
- GitHub Enterprise
- BitBucket
- BitBucket Server
- GitLab.com
- GitLab Self-Managed
- Hasicorp Vault [(coming soon)](https://coder.com/contact)

When users create/update workspaces, Coder will <a href="https://www.kapwing.com/e/631cf6a369c1ee00e55ff6ab" target="_blank">prompt users</a>
to authenticate with the provider if a valid token is not present.

```hcl
resource "coder_user_token" "github-enterprise" {
  type                = "github"
  host                = "https://github-enterprise.example.com"
  oauth_client_id     = var.github_client_id # via environment variable
  oauth_client_secret = var.github_client_secret # via environment variable

  add_coder_key  = true
  scopes         = ["read:user", "write:public_key", "write:gpg_key", "repo"]
}
```

> See the [Coder Terraform provider docs](#needs-link) for examples for each platform.

## Displaying Secrets

While you can inject secrets into the workspace via environment variables, you
can also show them in the Workspace UI with [`coder_metadata`](https://registry.terraform.io/providers/coder/coder/latest/docs/resources/metadata).

![secret UI](./images/secret-metadata-ui.png)

Can be produced with

```hcl
resource "twilio_iam_api_key" "api_key" {
  account_sid   = "ACXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX"
  friendly_name = "Test API Key"
}


resource "coder_metadata" "twilio_key" {
  resource_id = twilio_iam_api_key.api_key.id
  item {
    key = "secret"
    value = twilio_iam_api_key.api_key.secret
    sensitive = true
  }
}
```
