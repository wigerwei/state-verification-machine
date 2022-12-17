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
for i in range(8):
    G.add_node(i)

# Add directed edges
G.add_edge(0, 1)
G.add_edge(1, 5)
G.add_edge(3, 4)
G.add_edge(1, 3)
G.add_edge(6, 5)
G.add_edge(2, 7)

# Draw the graph
nx.draw(G, with_labels=True)
plt.show()
