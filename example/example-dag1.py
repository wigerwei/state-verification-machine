#!/usr/bin/env python
'''
@FileName    : example-dag1.py
@Description :
@Date        : 2022/12/16 23:44:16
@Author      : Wiger
@version     : 1.0
'''
# Import the required libraries
import networkx as nx
import matplotlib.pyplot as plt

# Create an empty graph with no nodes and no edges
g = nx.DiGraph()

# Add nodes to the graph
g.add_node(1)
g.add_node(2)
g.add_node(3)
g.add_node(4)
g.add_node(5)
g.add_node(6)
g.add_node(7)
g.add_node(8)
g.add_node(9)

# Add edges to the graph
g.add_edge(1, 2)
g.add_edge(1, 3)
g.add_edge(2, 4)
g.add_edge(3, 4)
g.add_edge(4, 5)
g.add_edge(5, 6)
g.add_edge(6, 1)
g.add_edge(7, 7)
g.add_edge(8, 9)
g.add_edge(9, 8)

# Draw the graph
nx.draw(g, with_labels=True)
plt.show()
