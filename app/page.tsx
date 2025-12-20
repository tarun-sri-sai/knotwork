import { handleSignup, handleLogin } from "./actions";

const Home = () => {
  return (
    <form className="flex flex-col items-center justify-center gap-20">
      <h2 className="text-lg">Welcome</h2>

      <div className="flex flex-col gap-4">
        <div className="field grid grid-cols-2 gap-4">
          <p>New to this?</p>
          <button formAction={handleSignup} type="submit" className="text-right">
            Signup
          </button>
        </div>

        <div className="field grid grid-cols-2 gap-4">
          <p>Already have an account?</p>
          <button formAction={handleLogin} type="submit" className="text-right">
            Login
          </button>
        </div>
      </div>
    </form>
  );
};

export default Home;
