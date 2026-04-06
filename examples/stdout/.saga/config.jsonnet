{
  audiences: [
    {
      name: "internal",
      description: "Internal team.",
      interest: "Technical changes, bug fixes, performance.",
      tone: "Direct and technical.",
      routes: [
        {
          name: "internal-stdout",
          courier: {
            name: "stdout",
            config: {},
          },
        },
      ],
    },
    {
      name: "external",
      description: "External users.",
      interest: "New features, improvements, important fixes.",
      tone: "Friendly and clear. No jargon.",
      routes: [
        {
          name: "external-stdout",
          courier: {
            name: "stdout",
            config: {
              prefix: "PUBLIC",
            },
          },
        },
      ],
    },
  ],
}
