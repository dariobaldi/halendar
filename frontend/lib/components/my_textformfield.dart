import 'package:flutter/material.dart';

class MyTextFormField extends StatelessWidget {
  final TextEditingController? controller;
  final String lableText;
  final bool obscureText;
  final Iterable<String>? hints;
  final void Function(BuildContext) onEnter;
  final double? width;
  final String? Function(String?)? validator;
  final void Function(PointerDownEvent)? onTapOutside;
  final Widget? suffixIcon;

  const MyTextFormField({
    super.key,
    required this.controller,
    required this.lableText,
    this.obscureText = false,
    required this.onEnter,
    this.hints = const <String>[],
    this.width = 300,
    this.validator,
    this.onTapOutside,
    this.suffixIcon,
  });

  @override
  Widget build(BuildContext context) {
    return Padding(
        padding: const EdgeInsets.symmetric(horizontal: 20.0),
        child: SizedBox(
          width: width,
          child: TextFormField(
            validator: validator?? (value) {
                  if (value == null || value.isEmpty) {
                    return 'La valeur ne peux pas être vide';
                  }
                  return null;
                },
            autofillHints: hints,
            onEditingComplete: () {onEnter(context);},
            controller: controller,
            obscureText: obscureText,
            decoration: InputDecoration(
              enabledBorder: OutlineInputBorder(
                borderSide:
                    BorderSide(color: Theme.of(context).colorScheme.outline),
              ),
              focusedBorder: OutlineInputBorder(
                borderSide:
                    BorderSide(color: Theme.of(context).colorScheme.primary),
              ),
              labelText: lableText,
              suffixIcon: suffixIcon,
            ),
            onTapOutside: onTapOutside,
          ),
        ));
  }
}
