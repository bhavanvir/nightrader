import React, { useState, useEffect } from "react";
import axios from "axios";

import FundsIcon from "../../assets/icons/FundsIcon";
import Clock from "./Clock";

export default function AccountInfo({ user, showAlert }) {
  const [balance, setBalance] = useState(0);

  let canadianDollar = new Intl.NumberFormat("en-CA", {
    style: "currency",
    currency: "CAD",
    minimumFractionDigits: 0,
  });

  const fetchWalletBalance = async () => {
    await axios
      .get("http://localhost/transaction/getWalletBalance", {
        withCredentials: true,
        headers: {
          token: localStorage.getItem("token"),
        },
      })
      .then(function (response) {
        setBalance(response.data.data.balance);
      })
      .catch(function (error) {
        showAlert(
          "error",
          "There was an error fetching your wallet balance. Please try again"
        );
      });
  };

  const handleClick = async () => {
    const fundsInput = document.getElementById("funds-modal-input").value;
    const fundsModal = document.getElementById("funds-modal");
    await axios
      .post(
        "http://localhost/transaction/addMoneyToWallet",
        {
          amount: parseInt(fundsInput),
        },
        {
          withCredentials: true,
          headers: {
            token: localStorage.getItem("token"),
          },
        }
      )
      .then(function (response) {
        showAlert("success", "Successfully added funds to your wallet!");
        fundsModal.close();
        fetchWalletBalance();
      })
      .catch(function (error) {
        showAlert(
          "error",
          "There was an error adding funds to your wallet. Please try again"
        );
      });
  };

  useEffect(() => {
    fetchWalletBalance();
    // eslint-disable-next-line
  }, []);

  return (
    <div className="mt-14 ">
      <div className="grid grid-cols-2 gap-6">
        <div className="card bg-base-300 shadow-xl">
          <div className="card-body">
            <h1 className="text-xl font-bold">Welcome back, {user.name}</h1>
            <h2 id="liveDateTime" className="text-xl">
              <Clock />
            </h2>
          </div>
        </div>
        <div>
          <div className="card bg-base-300 shadow-xl">
            <div className="card-body">
              <h1 className="text-xl font-bold">Account balance</h1>
              <div className="flex items-center justify-between">
                <h2 className="text-4xl">{canadianDollar.format(balance)}</h2>
                <button
                  className="btn"
                  onClick={() =>
                    document.getElementById("funds-modal").showModal()
                  }
                >
                  Add funds
                  <FundsIcon />
                </button>
              </div>

              <dialog id="funds-modal" className="modal">
                <div className="modal-box">
                  <h3 className="text-lg font-bold">
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
                          // Prevent non-numeric characters except for the negative sign
                          if (!/[0-9]/.test(event.key)) {
                            event.preventDefault();
                          } else if (event.key === "-") {
                            // Disable the button if the input is negative
                            document.getElementById(
                              "funds-modal-input"
                            ).disabled = true;
                          }
                        }}
                        onChange={(event) => {
                          // Prevent negative values from being entered
                          if (event.target.value < 0) {
                            event.target.value = 0;
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
