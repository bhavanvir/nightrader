import { Link } from "react-router-dom";

export default function Hero() {
  return (
    <div className="hero bg-base-200 min-h-screen">
      <div className="hero-content grid grid-rows-2 text-center">
        <div>
          <h1 className="text-5xl font-bold">
            Take control of your financial future.
          </h1>
          <h2 className="text-2xl">
            Sign up for early access to our all new trading platform.
          </h2>
        </div>

        <div className="justify-start">
          <Link to="/signin">
            <button className="btn btn-primary btn-wide btn-md m-4">
              Sign in
            </button>
          </Link>
          <Link to="/signup">
            <button className="btn btn-primary btn-outline btn-wide btn-md">
              Sign up
            </button>
          </Link>
        </div>
      </div>
    </div>
  );
}
