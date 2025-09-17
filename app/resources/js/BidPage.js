// document.querySelectorAll(".card-btn-add").forEach(btn => {
//     btn.addEventListener("click", function() {
//         const ServiceID = btn.getAttribute('btn-id');
//         const BidID = "1";
//         AddSoftwareDevServiceToBid(ServiceID, BidID);
//     })
// })

// let delBidBtn = document.getElementsByClassName("choose-btn-del")[0]
// delBidBtn.addEventListener("click", async function() {
//     const userID = "1";
//     const bidID = String(await GetUserActiveBid(userID));
//     console.log(bidID);
//     DeleteSoftwareDevBid(bidID);
// })

// async function GetUserActiveBid(userID) {
//     try {
//         const response = await fetch("http://localhost:80/api/findactivebid", {
//             method: "POST",
//             headers: {
//                 'Content-Type': 'application/json',
//             },
//             body: JSON.stringify({
//                 user: userID,
//             })
//         });
        
//         const data = await response.json();
//         return data["value"];
//     } catch (error) {
//         console.error("error:", error);
//     }
// }

// async function DeleteSoftwareDevBid(bidID) {
//     try {
//         const response = await fetch("http://localhost:80/api/deletebid", {
//             method: "POST",
//             headers: {
//                 'Content-Type': 'application/json',
//             },
//             body: JSON.stringify({
//                 bid: bidID,
//             })
//         });

//         const data = await response.json();
//         console.log(data);
//         window.location.href = "/services";
//     } catch (error) {
//         console.error("error:", error);
//     }
// }