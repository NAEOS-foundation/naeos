## naeos plugin

Manage NAEOS plugins

### Synopsis

Manage NAEOS plugins (install, uninstall, list, enable, disable, execute, info).

Example:
  naeos plugin list
  naeos plugin install ./my-plugin.so
  naeos plugin uninstall my-plugin
  naeos plugin enable my-plugin
  naeos plugin disable my-plugin
  naeos plugin info my-plugin
  naeos plugin execute my-plugin lint --params '{"file":"main.go"}'

### Options

```
  -h, --help                help for plugin
      --plugin-dir string   plugin directory (default "/home/codespace/.naeos/plugins")
```

### Options inherited from parent commands

```
      --dry-run                global dry-run mode: preview without writing to disk
      --output-format string   output format: json, yaml, table (default "table")
      --verbose                enable verbose logging
```

### SEE ALSO

* [naeos](naeos.md)	 - NAEOS CLI - Declarative Engineering Runtime
* [naeos plugin create](naeos_plugin_create.md)	 - Create a new plugin skeleton
* [naeos plugin disable](naeos_plugin_disable.md)	 - Disable a plugin
* [naeos plugin enable](naeos_plugin_enable.md)	 - Enable a plugin
* [naeos plugin execute](naeos_plugin_execute.md)	 - Execute a plugin action
* [naeos plugin info](naeos_plugin_info.md)	 - Show plugin information
* [naeos plugin init](naeos_plugin_init.md)	 - Scaffold a new plugin project
* [naeos plugin install](naeos_plugin_install.md)	 - Install a plugin from a .so file
* [naeos plugin list](naeos_plugin_list.md)	 - List installed plugins
* [naeos plugin search](naeos_plugin_search.md)	 - Search for plugins in the registry
* [naeos plugin test](naeos_plugin_test.md)	 - Test a plugin by loading, initializing, and checking health
* [naeos plugin uninstall](naeos_plugin_uninstall.md)	 - Uninstall a plugin

