def get_confirmation(prompt: str) -> bool:
    confirmation_prompt = f"{prompt} [Y/n] "
    confirmation = input(confirmation_prompt)
    return confirmation.lower() in {"y", "yes", ""}
