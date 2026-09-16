import 'package:halendar_front/components/my_textformfield.dart';
import 'package:halendar_front/const.dart';
import 'package:halendar_front/models/users.dart';
import 'package:flutter/material.dart';

class InsertUser extends StatefulWidget {
  const InsertUser({super.key});

  @override
  State<InsertUser> createState() => _InsertUserState();
}

class _InsertUserState extends State<InsertUser> {
  final _formKey = GlobalKey<FormState>();
  var user = User(
    id: uuidNil,
    createdAt: DateTime.now(),
    name: "",
    email: "",
    username: "",
    activated: false,
    accessLevel: 1,
    version: 0,
  );

  final TextEditingController _name = TextEditingController();
  final TextEditingController _username = TextEditingController();
  final TextEditingController _password = TextEditingController();
  bool showPassword = false;

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: SingleChildScrollView(
        child: Form(
          key: _formKey,
          child: Padding(
            padding: const EdgeInsets.all(15.0),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const SizedBox(height: 15),
                Align(child: Text("Ajouter un utilisateur")),
                const SizedBox(height: 10),
                MyTextFormField(
                  controller: _name,
                  lableText: "Nom",
                  onEnter: (context) {
                    _username.text = _name.text.replaceAll(" ", "_").toLowerCase();
                  },
                  onTapOutside: (value) {
                    _username.text = _name.text.replaceAll(" ", "_").toLowerCase();
                  },
                ),
                const SizedBox(height: 10),
                MyTextFormField(
                  controller: _username,
                  lableText: "Pseudo",
                  onEnter: (context) {
                    _username.text = _username.text.replaceAll(" ", "_").toLowerCase();
                  },
                  onTapOutside: (value) {
                    _username.text = _username.text.replaceAll(" ", "_").toLowerCase();
                  },
                ),
                const SizedBox(height: 5),
                  MyTextFormField(
                    controller: _password,
                    lableText: 'Mot de passe',
                    obscureText: !showPassword,
                    onEnter: (context){},
                    hints: const [AutofillHints.password],
                  ),
                const SizedBox(height: 5),
                Padding(
                  padding: const EdgeInsets.all(10.0),
                  child: Row(
                    mainAxisAlignment: MainAxisAlignment.end,
                    children: [
                      ActionChip(
                        onPressed: () {
                          Navigator.of(context).pop();
                        },
                        label: const Text("Annuler"),
                      ),
                      ActionChip(
                        onPressed: () {
                          user.name = _name.text;
                          user.username = _username.text;
                          registerUser(user, _password.text);
                          Navigator.of(context).pop();
                        },
                        label: const Text("Enregistrer"),
                      ),
                    ],
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}
