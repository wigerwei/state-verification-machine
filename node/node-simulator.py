#!/usr/bin/env python
'''
@FileName    : node-simulator.py
@Description : Node simulator in python.
@Date        : 2022/12/20 13:03:35
@Author      : Wiger
@version     : 1.0
'''

type Config struct {
	L1 L1EndpointSetup
	L2 L2EndpointSetup

	Driver driver.Config

	Rollup rollup.Config
}