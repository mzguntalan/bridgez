"use client";
import { Button } from "@/components/ui/button";
import React, { useState, useMemo, useRef, useEffect } from "react";
import { Textarea } from "@/components/ui/textarea";
import { SendIcon } from "lucide-react";
import { ForceGraph3D, ForceGraph2D } from "react-force-graph";
import { link, unwatchFile } from "fs";

function Bar(props: {
  value: string;
  onValueChange: (newValue: string) => void;
  onValueSubmit: (currentValue: string) => void;
}) {
  return (
    <div className="absolute inset-x-0 bottom-0 bg-white">
      <div className="bg-blue-100 px-10 py-2 text-center text-slate-500">
        Version 0.0.0
      </div>
      <div className=" px-10 py-5 flex items-center">
        <Textarea
          className="flex-1 resize-none "
          placeholder="Type your sentence here"
          value={props.value}
          onChange={(e) => {
            props.onValueChange(e.target.value);
          }}
        ></Textarea>
        <Button
          className="w-10 h-10 ms-2 rounded-full bg-blue-100"
          variant="ghost"
          size="icon"
          onClick={() => props.onValueSubmit(props.value)}
        >
          <SendIcon />
        </Button>
      </div>
    </div>
  );
}

function newGraph(): Graph {
  return { nodes: [], links: [] };
}
interface Node {
  id: string;
  name: string;
}
interface Link {
  source: string;
  target: string;
}
interface Graph {
  nodes: Node[];
  links: Link[];
}

function addNode(graph: Graph, node: Node): Graph {
  const index = graph.nodes.findIndex((v) => v.id === node.id);
  if (index !== -1) {
    return graph;
  }
  return {
    nodes: [...graph.nodes, node],
    links: graph.links,
  };
}

function addLink(graph: Graph, link: Link) {
  return {
    nodes: graph.nodes,
    links: [...graph.links, link],
  };
}

function getAllQuoted(sentence: string): string[] {
  const regexp = /「(.*?)」/g;
  const matches = [...sentence.matchAll(regexp)].map((m) => m[1]);

  return matches;
}

function getFocused(sentence: string): string[] {
  const regexp = /『(.*?)』/g;
  const matches = [...sentence.matchAll(regexp)].map((m) => m[1]);

  return matches;
}

function surrondWithQuotes(word: string): string {
  return `「${word}」`;
}

async function postWord(word: string): Promise<Response> {
  console.log(JSON.stringify({ id: word, name: word }));
  try {
    const response = await fetch("http://localhost:8080/word/add", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ id: word, name: word }),
    });

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    return response;
  } catch (error) {
    console.error("Fetch failed:", error);
    throw error;
  }
}

async function postLink(source: string, target: string): Promise<Response> {
  try {
    const response = await fetch("http://localhost:8080/link/add", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ source: source, target: target }),
    });

    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`);
    }

    return response;
  } catch (error) {
    console.error("Fetch failed:", error);
    throw error;
  }
}

export default function Home() {
  const [winWidth, winHeight] = [window.innerWidth, window.innerHeight];
  const [sentence, setSentence] = useState("");

  const [words, setWords] = useState<Node[] | undefined>(undefined);
  const [links, setLinks] = useState<Link[] | undefined>(undefined);
  const [myGraph, setMyGraph] = useState(newGraph());

  useEffect(() => {
    fetch("http://localhost:8080/words")
      .then((response) => response.json())
      .then((data) => {
        console.log("words", data);
        setWords(structuredClone(data));
        setMyGraph({
          nodes: structuredClone(data),
          links: myGraph.links,
        });
      });
  }, []);

  useEffect(() => {
    if (words === undefined) return;
    fetch("http://localhost:8080/links")
      .then((response) => response.json())
      .then((data) => {
        console.log("links", data);
        setLinks(structuredClone(data));
        setMyGraph({
          nodes: myGraph.nodes,
          links: structuredClone(data),
        });
      });
  }, [words]);

  return (
    <div className=" flex flex-col min-h-screen ">
      <div className="bg-gray-100 flex-1">
        <ForceGraph2D
          width={winWidth}
          height={winHeight}
          graphData={myGraph}
          linkCurvature={0.3}
          linkDirectionalArrowLength={5}
          nodeCanvasObjectMode={() => "after"}
          nodeCanvasObject={(node, ctx, globalScale) => {
            const label = node.name;
            const fontSize = 12 / globalScale;
            ctx.font = `${fontSize}px Sans-Serif`;
            ctx.textAlign = "center";
            ctx.textBaseline = "middle";
            ctx.fillStyle = "black"; //node.color;
            ctx.fillText(label, node.x, node.y + 10);
          }}
          onNodeClick={(node, event) => {
            setSentence(sentence + surrondWithQuotes(node.name));
          }}
        />
      </div>

      <Bar
        value={sentence}
        onValueChange={(e) => {
          setSentence(e);
          getAllQuoted(e);
          getFocused(e);
        }}
        onValueSubmit={(sentence) => {
          let targetWords = getAllQuoted(sentence);
          let focusedWord = getFocused(sentence)[0]; // todo: find checks

          if (focusedWord === undefined) {
            if (targetWords.length === 0) {
              alert("quote something please");
              return;
            }

            focusedWord = targetWords.at(-1) ?? "";
            targetWords = targetWords.slice(0, -1);
          }

          const nodes = [...targetWords, focusedWord].map((w) => {
            return {
              id: w,
              name: w,
            };
          });
          const links = targetWords.map((w) => {
            return {
              source: focusedWord,
              target: w,
            };
          });

          let graph = myGraph;
          for (const node of nodes) {
            graph = addNode(graph, node);
            postWord(node.id);
          }

          for (const link of links) {
            graph = addLink(graph, link);
            postLink(link.source, link.target);
          }

          setMyGraph(graph);
        }}
      />
    </div>
  );
}
