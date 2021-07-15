import React, { useState } from "react";
import useSWR from "swr";
import logo from "./logo.svg";

const fetcher = (url: string) => fetch(url).then((res) => res.json());

function App() {
  const { data } = useSWR<number[]>("/api/data", fetcher, {
    refreshInterval: 1000,
  });

  const d = data ?? [];

  return (
    <div>
      <button
        onClick={async () => {
          await fetch("/api/data", { method: "POST" });
        }}
      >
        Send task
      </button>
      <h1>Api results</h1>
      <div className="flex-col flex">
        {d.map((o) => (
          <p>{o}</p>
        ))}
      </div>
    </div>
  );
}

export default App;
