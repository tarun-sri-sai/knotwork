import { handleLogin } from "./actions";
import ErrorMessage from "@/components/ErrorMessage";

const Login = async ({ searchParams }: { searchParams: Promise<Record<string, string | string[] | undefined>> }) => {
  const resolvedSearchParams = await searchParams;
  const error = resolvedSearchParams.error;

  return (
    <form className="flex flex-col items-center justify-center gap-20">
      <h2 className="text-lg">Please enter the email and password</h2>

      <div className="flex flex-col gap-4">
        <div className="field grid grid-cols-2 gap-4">
          <label htmlFor="email">Email: </label>
          <input id="email" type="email" />
        </div>

        <div className="field grid grid-cols-2 gap-4">
          <label htmlFor="password">Password: </label>
          <input id="password" type="password" />
        </div>

        <div className="field grid grid-cols-2 gap-4">
          <button formAction={handleLogin} type="submit" className="text-right">
            Login
          </button>
        </div>
      </div>

      {error && <ErrorMessage message={error} />}
    </form>
  );
};

export default Login;
