{
  tales: [
    {
      audience: "engineering",
      title: "OAuth2 PKCE replaces /auth/token",
      text: |||
        The legacy /auth/token endpoint is **deprecated** and will be
        removed in 0.3.0. All integrations must migrate to the new
        OAuth2 PKCE flow via /oauth/authorize and /oauth/token.

        ## Migration

        See [the migration guide](https://example.com/docs/oauth-migration)
        in the wiki for step-by-step instructions and example code.
      |||,
    },
    {
      audience: "company",
      title: "Login security upgrade",
      text: |||
        We've upgraded our login security to a modern standard called
        **OAuth2 PKCE**. This makes all integrations safer against
        token theft.

        _No action is needed for most users_, but custom integrations
        will need to update before 0.3.0.
      |||,
    },
  ],
}
