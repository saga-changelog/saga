{
  tales: [
    {
      audience: "engineering",
      title: "API rate limits are now per-account",
      text: |||
        Rate limiting is now enforced at the account level instead of
        globally. The default is **1000 requests per minute per account**.

        ## What to watch for

        Clients that relied on the shared global pool will start seeing
        429 responses sooner if they exceed their per-account quota.
        Check the X-RateLimit response headers for current usage. See
        [the rate-limiting docs](https://example.com/docs/rate-limits)
        for the full header reference.
      |||,
    },
  ],
}
