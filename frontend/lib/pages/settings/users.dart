import 'package:halendar_front/const.dart';
import 'package:halendar_front/models/users.dart';
import 'package:halendar_front/pages/settings/insert/user.dart';
import 'package:halendar_front/services/auth.dart';
import 'package:flutter/material.dart';
import 'package:provider/provider.dart';

class UsersPage extends StatelessWidget {
  final UsersModel m;
  const UsersPage({super.key, required this.m});

  void toggleActivation(BuildContext context, User user) {
    showDialog(
      context: context,
      builder:
          (context) => AlertDialog(
            title: Text(
              "${user.activated ? "Deactivate" : "Activate"} ${user.name}'s account?",
            ),
            actions: [
              TextButton(
                onPressed: () {
                  Navigator.of(context).pop();
                },
                child: const Text("Cancel"),
              ),
              TextButton(
                onPressed:
                    (AuthService.instance.accessLevel <= user.accessLevel)
                        ? null
                        : () {
                          user.activated = !user.activated;
                          activateUser(user);
                          Navigator.of(context).pop();
                        },
                child: const Text("Save"),
              ),
            ],
          ),
    );
  }

  void changeLevel(BuildContext context, User user) {
    showDialog(
      context: context,
      builder:
          (context) => AlertDialog(
            title: Text("Change access level"),
            content:
                (AuthService.instance.accessLevel <= user.accessLevel)
                    ? Text("Unavailable")
                    : Column(
                      spacing: 5,
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        LevelButton(user: user, level: 1),
                        LevelButton(user: user, level: 5),
                        LevelButton(user: user, level: 10),
                      ],
                    ),
            actions: [
              TextButton(
                onPressed: () {
                  Navigator.of(context).pop();
                },
                child: const Text("Cancel"),
              ),
            ],
          ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text("Users"),
        actions: [
          ElevatedButton(
            onPressed: () {
              Navigator.push(
                context,
                MaterialPageRoute(builder: (context) => InsertUser()),
              );
            },
            child: Icon(Icons.add),
          ),
        ],
      ),
      body: Column(
        children: [
          // const SizedBox(height: 15),
          if (m.fetching) const Center(child: CircularProgressIndicator()),
          if (!m.fetching)
            Expanded(
              child: ListView.builder(
                padding: const EdgeInsets.only(top: 8),
                itemCount: m.users.length,
                itemBuilder: (context, index) {
                  return Padding(
                    padding: const EdgeInsets.only(left: 8, right: 8, top: 5),
                    child: ListTile(
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(12),
                      ),
                      tileColor: Theme.of(
                        context,
                      ).colorScheme.primary.withAlpha(90),
                      title: Text(m.users[index].name),
                      subtitle: Text(
                        accessLevelName(m.users[index].accessLevel),
                      ),
                      leading: IconButton(
                        onPressed: () {
                          changeLevel(context, m.users[index]);
                        },
                        icon: accessLevelIcon(m.users[index].accessLevel),
                      ),
                      trailing: Row(
                        spacing: 3,
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          Text(
                            m.users[index].activated ? "Active" : "Inactive",
                          ),
                          Icon(
                            Icons.circle,
                            color:
                                m.users[index].activated
                                    ? Colors.green
                                    : Colors.red,
                          ),
                        ],
                      ),
                      onTap: () {
                        toggleActivation(context, m.users[index]);
                      },
                    ),
                  );
                },
              ),
            ),
        ],
      ),
    );
  }
}

class LevelButton extends StatelessWidget {
  final User user;
  final int level;
  const LevelButton({super.key, required this.user, required this.level});

  @override
  Widget build(BuildContext context) {
    return FilledButton(
      onPressed:
          (AuthService.instance.accessLevel <= user.accessLevel)
              ? null
              : () {
                changeUserLevel(user, level);
                Navigator.of(context).pop();
              },
      child: Row(
        mainAxisSize: MainAxisSize.min,
        spacing: 3,
        children: [accessLevelIcon(level), Text(accessLevelName(level))],
      ),
    );
  }
}

// DEPENDENCY INJECTION

class UsersScreen extends StatelessWidget {
  const UsersScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return MultiProvider(
      providers: [ChangeNotifierProvider(create: (_) => UsersModel())],
      child: const InitModel(),
    );
  }
}

class InitModel extends StatefulWidget {
  const InitModel({super.key});

  @override
  State<InitModel> createState() => _InitModelState();
}

class _InitModelState extends State<InitModel> {
  late UsersModel m;
  @override
  void initState() {
    super.initState();
    m = Provider.of<UsersModel>(context, listen: false);
    m.init();
  }

  @override
  void dispose() {
    m.end();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return const SelectionArea(child: ModelConsumer());
  }
}

class ModelConsumer extends StatelessWidget {
  const ModelConsumer({super.key});

  @override
  Widget build(BuildContext context) {
    return Consumer<UsersModel>(
      builder: (context, m, _) {
        return UsersPage(m: m);
      },
    );
  }
}
