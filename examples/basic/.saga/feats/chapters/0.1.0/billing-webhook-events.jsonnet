{
  tales: [
    {
      audience: "billing",
      title: "Real-time webhooks for billing state changes",
      text: |||
        New webhook events are now fired for all billing state changes:
        invoice.created, invoice.paid, invoice.overdue, and
        subscription.changed. See [BILL-892](https://example.com/tickets/BILL-892)
        for the full event schema.

        Finance dashboards **should subscribe** to these for real-time
        reconciliation.
      |||,
    },
    {
      audience: "engineering",
      title: "Billing events on the event bus",
      text: |||
        Billing state changes now emit webhook events via the existing
        event bus. New event types: invoice.created, invoice.paid,
        invoice.overdue, subscription.changed.

        The schema lives in the events package. Consumers can filter by
        event type on the /webhooks/subscribe endpoint. See
        [the events package docs](https://example.com/docs/events) for
        examples.
      |||,
    },
    {
      audience: "company",
      title: "Instant billing notifications",
      text: |||
        Our billing system now sends **real-time notifications** when
        invoices are created, paid, or become overdue.

        This means finance gets _instant visibility_ into billing status
        without checking dashboards manually.
      |||,
    },
  ],
}
