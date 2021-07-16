import React, { useEffect, useRef, useState } from "react";
import useSWR from "swr";
import logo from "./logo.svg";

const fetcher = (url: string) => fetch(url).then((res) => res.json());

function App() {
  const { data } = useSWR<number[]>("/api/data", fetcher, {
    refreshInterval: 500,
  });

  const d = data ?? [];

  const [sendTasksEvery, setSendTasksEvery] = useState(0);

  useEffect(() => {
    let t: number;
    const fn = async () => {
      if (sendTasksEvery) {
        await fetch("/api/data", { method: "POST" });
        t = setTimeout(async () => {
          await fn();
        }, sendTasksEvery);
      }
    };
    if (sendTasksEvery) {
      fn();
    }
    return () => {
      clearTimeout(t);
    };
  }, [sendTasksEvery]);

  return (
    <div>
      <button
        onClick={async () => {
          await fetch("/api/data", { method: "POST" });
        }}
      >
        Send task
      </button>
      <input
        value={sendTasksEvery / 1000}
        step="10"
        placeholder="Send tasks every n second(s)"
        type="number"
        onChange={(e) => {
          const val = e.target.value;
          setSendTasksEvery(Math.floor(Number(val) * 1000));
        }}
      />
      <button></button>
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
