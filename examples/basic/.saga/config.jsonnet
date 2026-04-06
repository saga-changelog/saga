{
  audiences: [
    {
      name: "engineering",
      description: "Internal engineering team.",
      interest: "API changes, schema migrations, breaking changes, performance improvements.",
      tone: "Technical. Reference APIs, schemas, and endpoints.",
      routes: [
        {
          name: "engineering-slack",
          courier: {
            name: "slack-legacy",
            config: {
              channel: "#engineering",
            },
          },
        },
      ],
    },
    {
      name: "company",
      description: "Company-wide announcements.",
      interest: "Product improvements, major changes, new integrations.",
      tone: "Friendly and clear. Accessible to non-technical readers.",
      routes: [
        {
          name: "company-basecamp",
          courier: {
            name: "basecamp-messageboard",
            config: {
              project_id: "12345678",
              message_board_id: "87654321",
            },
          },
        },
      ],
    },
    {
      name: "billing",
      description: "Internal finance department.",
      interest: "Billing cycle changes, invoice formats, tax handling, payment terms.",
      tone: "Direct and technical. Reference ticket numbers.",
      routes: [
        {
          name: "billing-slack",
          courier: {
            name: "slack-legacy",
            config: {
              channel: "#finance-updates",
            },
          },
        },
      ],
    },
  ],
}
