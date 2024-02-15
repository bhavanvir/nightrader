import React, { useState, useEffect } from "react";
import axios from "axios";

import FundsIcon from "../../assets/icons/FundsIcon";
import Clock from "./Clock";

export default function Hero({ user, showAlert }) {
  const [balance, setBalance] = useState(0);

  const fetchWalletBalance = async () => {
    await axios
      .get("http://localhost:5000/getWalletBalance", {
        withCredentials: true,
      })
      .then(function (response) {
        setBalance(response.data.data.balance);
      })
      .catch(function (error) {
        showAlert("error", error.response.data.message);
      });
  };

  const handleClick = async () => {
    const funds = document.getElementById("funds-modal-input").value;
    await axios
      .post(
        "http://localhost:5000/addMoneyToWallet",
        {
          amount: parseInt(funds),
        },
        {
          withCredentials: true,
        }
      )
      .then(function (response) {
        showAlert(
          "success",
          "Successfully added funds to your wallet! Refresh your page to see the updated balance"
        );
      })
      .catch(function (error) {
        showAlert("error", error.response.data.message);
      });
  };

  useEffect(() => {
    fetchWalletBalance();
  }, []);

  return (
    <div className="mt-14 ">
      <div className="grid grid-cols-2 gap-6">
        <div className="card bg-base-300 shadow-xl">
          <div className="card-body">
            <h1 className="text-xl font-bold">Welcome back, {user.name}</h1>
            <h2 id="liveDateTime" className="text-lg">
              <Clock />
            </h2>
          </div>
        </div>
        <div>
          <div className="card bg-base-300 shadow-xl">
            <div className="card-body">
              <h1 className="text-xl font-bold">Account balance</h1>
              <h2 className="text-lg">${balance}</h2>
              <button
                className="btn"
                onClick={() =>
                  document.getElementById("funds-modal").showModal()
                }
              >
                Add funds
                <FundsIcon />
              </button>

              <dialog id="funds-modal" className="modal">
                <div className="modal-box">
                  <h3 className="font-bold text-lg">
                    How much would you like to add?
                  </h3>
                  <div className="grid grid-cols-2 gap-4 py-4">
                    <div>
                      <input
                        id="funds-modal-input"
                        type="number"
                        placeholder="Enter an amount"
                        className="input input-bordered w-full"
                        onKeyPress={(event) => {
                          if (!/[0-9]/.test(event.key)) {
                            event.preventDefault();
                          }
                        }}
                      />
                    </div>
                    <div>
                      <button className="btn" onClick={handleClick}>
                        Add
                      </button>
                    </div>
                  </div>
                </div>
                <form method="dialog" className="modal-backdrop">
                  <button>close</button>
                </form>
              </dialog>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
