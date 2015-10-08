from yum.plugins import PluginYumExit, TYPE_CORE, TYPE_INTERACTIVE
from yum import config
import json, socket, urllib2, urllib

requires_api_version = '2.3'
plugin_type = (TYPE_CORE, TYPE_INTERACTIVE)


#(name, epoch, version, release, arch)
#nevra = (transaction.name, transaction.epoch, transaction.version,transaction.release,transaction.arch)


def getfiles(name, pkgs):
    for p in pkgs:
        try:
            if name == p.name:
                return (p.version, p.filelist)
        except:
            print name, "failed", len(pkgs)
            
def config_hook(conduit):
# This is how the docs at http://yum.baseurl.org/wiki/WritingYumPlugins recommend setting configuration
# but it isn't working so I copied the technique from the other plugins.
#    config.YumConf.pipevizurl = config.UrlOption(default="http://pipeserv:2309/", schemes=('http', 'https'))
    global pipeviz_url
    pipeviz_url = conduit.confString('main', 'pipeviz_url', default="http://pipeserv:2309/")



def posttrans_hook(conduit):
#   import pdb; pdb.set_trace()
    pkgs = conduit.getPackages()
    file_collections = []
    hostname = socket.getfqdn()
    libraries = list()
    with open("/tmp/conduit.json", "w") as f:
        for transaction in conduit.getTsInfo():
            file_collections.append(getfiles(transaction.name, pkgs))   
            for dep in transaction.depends_on:
                file_collections.append(getfiles(dep.name, pkgs))
        logic_states = {"logic-states":list()}
#do libraries first, disabled for now
#        for version, filelist in file_collections:
#            for file_name in filelist:
#                if file_name.find(".so") != -1:
#                    state = {"type":"library", 
#                        "id":
#                            {"version":version},
#                        "environment":
#                            {"address":
#                                {"hostname":hostname}
#                            },
#                    }
#                    libraries.append(file_name)
#                    state["path"] = file_name
#                    logic_states["logic-states"].append(state)

#then binaries
        for version, filelist in file_collections:
            for file_name in filelist:
                if file_name.find("bin/") != -1:
                    state = {"type":"binary", 
                        "id":
                            {"version":version},
                        "environment":
                            {"address":
                                {"hostname":hostname}
                            },
                    }
                    if libraries:
                        state["libraries"] = libraries
                    state["path"] = file_name
                    logic_states["logic-states"].append(state)
        #Dump a copy into /tmp for debugging.
        json.dump(logic_states, f, indent=2)
        req = urllib2.Request(pipeviz_url, data=json.dumps(logic_states))
        urllib2.urlopen(req)

# vim: set tabstop=4 shiftwidth=4 expandtab
