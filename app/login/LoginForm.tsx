import { login } from "./actions";
import Button from "@/components/Button";
import Input from "@/components/Input";

const LoginForm = () => {
  return (
    <form action={login} className="flex flex-col gap-4">
      <Input name="email" type="email" placeholder="Email" />
      <Input name="password" type="password" placeholder="Password" />
      <Button type="submit">Login</Button>
    </form>
  );
};

export default LoginForm;
