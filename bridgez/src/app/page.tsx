"use client";
import { Button } from "@/components/ui/button";
import React, { useState, useMemo, useRef } from "react";
import { Textarea } from "@/components/ui/textarea";
import { SendIcon } from "lucide-react";
import { ForceGraph3D, ForceGraph2D } from "react-force-graph";

function genRandomTree(N = 300, reverse = false) {
  return {
    nodes: [...Array(N).keys()].map((i) => ({ id: i, name: "ありがとう" })),
    links: [...Array(N).keys()]
      .filter((id) => id)
      .map((id) => ({
        [reverse ? "target" : "source"]: id,
        [reverse ? "source" : "target"]: Math.round(Math.random() * (id - 1)),
      })),
  };
}

function Bar(props: {
  value: string;
  onValueChange: (newValue: string) => void;
  onValueSubmit: (currentValue: string) => void;
}) {
  return (
    <div className="absolute inset-x-0 bottom-0 bg-white">
      <div className="bg-blue-100 px-10 py-2">Toolbar</div>
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

function newGraph() {
  return { nodes: [], links: [] };
}

function addNode(graph, node) {
  return {
    nodes: [...graph.nodes, node],
    links: graph.links,
  };
}

function addLink(graph, link) {
  return {
    nodes: graph.nodes,
    links: [...graph.links, link],
  };
}

function addLink2(graph, nodeA, nodeB) {
  return {
    nodes: graph.nodes,
    links: [
      ...graph.links,
      {
        source: nodeA.id,
        target: nodeB.id,
      },
    ],
  };
}

function getAllQuoted(sentence: string): string[] {
  const regexp = /「(.*?)」/g;
  const matches = [...sentence.matchAll(regexp)].map((m) => m[1]);

  console.log(matches);
  return matches;
}

function getFocused(sentence: string): string[] {
  const regexp = /『(.*?)』/g;
  const matches = [...sentence.matchAll(regexp)].map((m) => m[1]);

  console.log(matches);
  return matches;
}

export default function Home() {
  // const myGraph = {
  //   nodes: [0, 1, 2, 3, 4, 5, 6, 7].map((i) => {
  //     return { id: i, name: "Node: 猫ちゃん" + String(i) };
  //   }),

  //   links: [
  //     [0, 1],
  //     [1, 2],
  //     [2, 1],
  //     [0, 2],
  //     [0, 3],
  //     [0, 3],
  //     [0, 4],
  //     [4, 5],
  //     [4, 7],
  //     [5, 6],
  //     [7, 6],
  //     [6, 7],
  //   ].map((p, i) => {
  //     return {
  //       source: p[0],
  //       target: p[1],
  //       name: "Some Link " + String(p[0]) + `<${i}>` + String(p[1]),
  //     };
  //   }),
  // };

  const [myGraph, setMyGraph] = useState(newGraph());

  const theGraph = useMemo(() => myGraph, []);

  const rndGraph = genRandomTree(1_000);

  // const [winWidth, winHeight] = useWindowSize();
  const [winWidth, winHeight] = [window.innerWidth, window.innerHeight];
  const [sentence, setSentence] = useState("");

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
            setSentence(node.name);
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
          const targetWords = getAllQuoted(sentence);
          const focusedWord = getFocused(sentence)[0]; // todo: find checks

          console.log("targetWords", targetWords);
          console.log("focusedWord", focusedWord);

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
          }

          for (const link of links) {
            graph = addLink(graph, link);
          }

          console.log("graph", graph);

          setMyGraph(graph);
        }}
      />
    </div>
  );
}
