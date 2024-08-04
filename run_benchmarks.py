
from mininet.net import Mininet
from mininet.topo import Topo
from mininet.log import setLogLevel, info
from mininet.util import pmonitor
from subprocess import STDOUT

from dataclasses import dataclass
from typing import *
import tomllib
from operator import mul
from functools import reduce
from time import sleep

DEFAULT_COMMON_ARGS = {'executable': './benchmark', 'insecure': True, 'q': True, 'v': True, 'fec': [True, False], 'fecScheme': 'rs', 'trace': True}
DEFAULT_SERVER_ARGS = {'s': True}
DEFAULT_CLIENT_ARGS = {}

@dataclass
class NetConfig:
	loss: int
	bandwidth: int
	delay: int

@dataclass
class FinalConfig:
	net_config: NetConfig
	server_args: List[str]
	client_args: List[str]
	testcase_name: str

def run_mininet(config: FinalConfig):
	topo = Topo()
	server_s = topo.addHost('server')
	client_s = topo.addHost('client')
	switch = topo.addSwitch('s0')
	link1 = topo.addLink(client_s, switch, loss=config.net_config.loss, bandwidth=config.net_config.bandwidth, delay=config.net_config.delay)
	link2 = topo.addLink(server_s, switch)
	info('topology configured\n')
	
	net = Mininet(topo=topo, waitConnected=True)
	net.start()
	server = net.get(server_s)
	client = net.get(client_s)
	info('mininet started\n')

	popens = dict()
	last = net.hosts[ -1 ]
	url = config.client_args[-1]
	config.client_args[-1] = f"https://{server.IP()}:6121/{url}"
	pserver = server.popen(*config.server_args, stderr=STDOUT)
	pclient = client.popen(*config.client_args, stderr=STDOUT)
	info(f'processes started\n{config.server_args=}\n{config.client_args=}\n')

	# for host, line in pmonitor( {client: pclient, server:pserver} ):
	for host, line in pmonitor( {client: pclient} ):
		if host:
			info( "<%s>: %s" % ( host.name, line ) )

	# pclient.wait()
	sleep(1)
	pserver.terminate()


	net.stop()
	info('mininet stopped\n')

def convert_dict_to_args(data, pre=None, post=None):
	args = []
	if pre is not None:
		args.append(pre)
	for flag, val in data.items():
		if val == False:
			continue
		if val == True:
			args.append(f'-{flag}')
		else:
			args.append(f'-{flag}')
			args.append(f'{val}')
	if post is not None:
		args.append(post)
	return args

def single_to_list(item):
	if isinstance(item, list):
		return item
	else:
		return [item]

def second(tup):
	_,x = tup
	return x

def config_iterations(config_raw):
	config = {k: single_to_list(v) for k,v in config_raw.items()}
	digits = [(k, len(v)) for k, v in config.items()]
	maximum = reduce(mul, map(second, digits), 1)
	for i in range(maximum):
		info(f"{i=}, {maximum=}, {config=}\n")
		res_conf = {}
		index = i
		for key, count in digits:
			key_idx = index % count
			index //= count
			res_conf[key] = config[key][key_idx]
		yield res_conf

def generate_configs(toml_config, name):
	common_config = dict(DEFAULT_COMMON_ARGS.copy(), **toml_config.get('common', dict()))
	client_config = dict(DEFAULT_CLIENT_ARGS.copy(), **toml_config.get('client', dict()))
	server_config = dict(DEFAULT_SERVER_ARGS.copy(), **toml_config.get('server', dict()))
	net_config = toml_config.get('net', dict())
	for common in config_iterations(common_config):
		for client in config_iterations(client_config):
			for server in config_iterations(server_config):
				for net in config_iterations(net_config):
					client_conf_instance = dict(common.copy(), **client)
					server_conf_instance = dict(common.copy(), **server)

					net_conf_instance = NetConfig(
						loss= net.get('loss', 0),
						bandwidth= net.get('bandwidth', None),
						delay= net.get('delay', 0),
					)

					# convert to string form
					url = client_conf_instance.pop('url')
					executable = client_conf_instance.pop('executable')
					client_args = convert_dict_to_args(client_conf_instance, executable, url)
					executable = server_conf_instance.pop('executable')
					server_args = convert_dict_to_args(server_conf_instance, executable)
					yield FinalConfig(net_conf_instance, server_args, client_args, name)




def run_benchmarks():
	conf = tomllib.load(open('benchmark_config.toml', 'rb'))
	for testcase, testcase_conf in conf.items():
		for config in generate_configs(testcase_conf, testcase):
			run_mininet(config)

setLogLevel( 'info' )
run_benchmarks()
