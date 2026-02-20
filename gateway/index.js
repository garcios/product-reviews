const { ApolloServer } = require('apollo-server');
const { ApolloGateway, IntrospectAndCompose, RemoteGraphQLDataSource } = require('@apollo/gateway');

class MonitoredDataSource extends RemoteGraphQLDataSource {
    willSendRequest({ request, context }) {
        console.log(`[Apollo Gateway] Routing request to subgraph: ${this.url}`);
        // You can also log the query itself:
        // console.log(request.query);
    }
}

const gateway = new ApolloGateway({
    supergraphSdl: new IntrospectAndCompose({
        subgraphs: [
            { name: 'products', url: 'http://localhost:4001/query' },
            { name: 'users', url: 'http://localhost:4002/query' },
            { name: 'reviews', url: 'http://localhost:4003/query' }
        ],
    }),
    buildService({ name, url }) {
        return new MonitoredDataSource({ url });
    }
});

const server = new ApolloServer({
    gateway,
    subscriptions: false,
});

server.listen({ port: 4000 }).then(({ url }) => {
    console.log(`🚀 Gateway ready at ${url}`);
});
