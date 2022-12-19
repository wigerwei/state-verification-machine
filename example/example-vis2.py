#!/usr/bin/env python
'''
@FileName    : example2.py
@Description : A example of DAG graph visualization.
@Date        : 2022/12/14 17:34:27
@Author      : Wiger
@version     : 1.0
'''
# Import the required libraries
import networkx as nx
import matplotlib.pyplot as plt

# Create an empty graph
G = nx.DiGraph()

# Add 10 nodes
for i in range(10):
    G.add_node(i)

# Add directed edges
G.add_edge(0, 1)
G.add_edge(1, 3)
G.add_edge(5, 8)
G.add_edge(7, 9)
G.add_edge(4, 7)
G.add_edge(1, 6)
G.add_edge(1, 3)
G.add_edge(1, 3)

# Draw the graph
nx.draw(G, with_labels=True)
plt.show()
