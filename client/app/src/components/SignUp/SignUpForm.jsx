import React, { useState } from "react";
import LogoIcon from "../../assets/icons/LogoIcon";
import { Link } from "react-router-dom";

export default function SignUpForm() {
  const [formData, setFormData] = useState({
    name: "",
    username: "",
    password: "",
  });

  const [fieldValidity, setFieldValidity] = useState({
    name: true,
    username: true,
    password: true,
  });

  const [errorMessages, setErrorMessages] = useState({
    name: "",
    username: "",
    password: "",
  });

  const handleInputChange = (e) => {
    const { name, value } = e.target;
    setFormData((prevData) => ({
      ...prevData,
      [name]: value,
    }));

    setErrorMessages((prevErrors) => ({
      ...prevErrors,
      [name]: "",
    }));
  };

  const validateFields = () => {
    const nameRegex = /^[a-zA-Z\s]+$/;
    const usernameRegex = /^[a-zA-Z0-9_]+$/;
    const passwordRegex = /^(?=.*[A-Za-z])(?=.*\d)[A-Za-z\d@$!%*#?&]{8,}$/;

    const newFieldValidity = {
      name: nameRegex.test(formData.name.trim()),
      username: usernameRegex.test(formData.username.trim()),
      password: passwordRegex.test(formData.password.trim()),
    };

    setFieldValidity(newFieldValidity);

    const newErrorMessages = {
      name: newFieldValidity.name ? "" : "Please enter a valid name",
      username: newFieldValidity.username
        ? ""
        : "Username must contain only letters, numbers, and underscores",
      password: newFieldValidity.password
        ? ""
        : "Password must be at least 8 characters long and include at least one letter and one number",
    };

    setErrorMessages(newErrorMessages);

    return Object.values(newFieldValidity).every((isValid) => isValid);
  };

  const handleSubmit = (e) => {
    e.preventDefault();

    const isValid = validateFields();

    if (isValid) {
      console.log("Form Data:", formData);
      // Add logic to send form data to the database
    }
  };

  return (
    <div className="relative flex h-screen flex-col justify-center overflow-hidden">
      <div className="border-primary m-auto w-full rounded-md border p-6 shadow-md ring-2 ring-gray-800/50 lg:max-w-lg">
        <div className="flex justify-center">
          <Link to="/">
            <a href="/" className="btn btn-ghost text-xl">
              <LogoIcon />
              <span className="text-3xl font-semibold">Nightrader</span>
            </a>
          </Link>
        </div>
        <form className="space-y-4" onSubmit={handleSubmit}>
          <div>
            <label htmlFor="name" className="name">
              <span className="label-text text-base">Name</span>
            </label>
            <input
              type="text"
              id="name"
              name="name"
              placeholder="Name"
              className={`input input-bordered w-full ${
                fieldValidity.name ? "" : "border-red-500"
              }`}
              value={formData.name}
              onChange={handleInputChange}
            />
            {errorMessages.name && (
              <p className="text-sm text-red-500">{errorMessages.name}</p>
            )}
          </div>
          <div>
            <label htmlFor="username" className="username">
              <span className="label-text text-base">Username</span>
            </label>
            <input
              type="text"
              id="username"
              name="username"
              placeholder="Username"
              className={`input input-bordered w-full ${
                fieldValidity.username ? "" : "border-red-500"
              }`}
              value={formData.username}
              onChange={handleInputChange}
            />
            {errorMessages.username && (
              <p className="text-sm text-red-500">{errorMessages.username}</p>
            )}
          </div>
          <div>
            <label htmlFor="password" className="password">
              <span className="label-text text-base">Password</span>
            </label>
            <input
              type="password"
              id="password"
              name="password"
              placeholder="Enter Password"
              className={`input input-bordered w-full ${
                fieldValidity.password ? "" : "border-red-500"
              }`}
              value={formData.password}
              onChange={handleInputChange}
            />
            {errorMessages.password && (
              <p className="text-sm text-red-500">{errorMessages.password}</p>
            )}
          </div>
          <div>
            <button type="submit" className="btn btn-block btn-primary">
              Sign up
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
