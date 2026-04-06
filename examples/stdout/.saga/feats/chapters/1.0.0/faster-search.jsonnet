{
  tales: [
    {
      audience: "internal",
      title: "Search endpoint rewired onto an inverted index",
      text: |||
        Replaced the sequential scan on the search endpoint with an
        inverted index. **P99 latency** dropped from 800ms to 60ms.

        See [the perf notes](https://example.com/docs/search-perf) for
        the benchmark methodology.
      |||,
    },
    {
      audience: "external",
      title: "Search is now much faster",
      text: |||
        Search is now _significantly faster_. Results appear almost
        instantly, even for large queries.
      |||,
    },
  ],
}
