#!/usr/bin/env python
'''
@FileName    : partition.py
@Description : DAG could be partitioned into two independent subsets.
@Date        : 2022/12/14 17:42:53
@Author      : Wiger
@version     : 1.0
'''
def divide_into_independent_subsets(graph):
  # Create a reversed copy of the graph
  reversed_graph = {vertex: set() for vertex in graph}
  for vertex, edges in graph.items():
    for neighbor in edges:
      reversed_graph[neighbor].add(vertex)

  # Initialize the stack and visited set
  stack = []
  visited = set()

  # Depth-first search on the reversed graph to get the finish times of the vertices
  def dfs(vertex):
    visited.add(vertex)
    for neighbor in reversed_graph[vertex]:
      if neighbor not in visited:
        dfs(neighbor)
    stack.append(vertex)

  for vertex in reversed_graph:
    if vertex not in visited:
      dfs(vertex)

  # Initialize the strongly connected components
  sccs = []

  # Clear the visited set
  visited.clear()

  # Depth-first search on the original graph to find the strongly connected components
  def dfs(vertex, scc):
    visited.add(vertex)
    scc.add(vertex)
    for neighbor in graph[vertex]:
      if neighbor not in visited:
        dfs(neighbor, scc)

  while stack:
    vertex = stack.pop()
    if vertex not in visited:
      scc = set()
      dfs(vertex, scc)
      sccs.append(scc)

  # Return the strongly connected components
  return sccs

graph = {
  '1': {'2', '3'},
  '2': {'4'},
  '3': {'4'},
  '4': {'5'},
  '5': {'6'},
  '6': {'1'},
  '7': {'7'},
  '8': {'9'},
  '9': {'8'}
}

sccs = divide_into_independent_subsets(graph)
print(sccs)
