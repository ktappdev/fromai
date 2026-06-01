// module.exports = {
//     apps: [
//         {
//             name: "lyricut-front",
//             script: "pnpm",
//             args: "preview",
//             interpreter: "pnpm",
//             env: {
//                 NODE_ENV: "production",
//                 PORT: "3000",
//             },
//         },
//     ],
// };
module.exports = {
  apps: [
    {
      name: 'fromai-frontwnd',
      script: './build/index.js',
      max_memory_restart: '1G',
      env: {
        NODE_ENV: 'production',
        PORT: 3001
      }
    }
  ]
};

