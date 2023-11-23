# kafka-doc

kafka-doc eases searching and reading [Apache Kafka](https://kafka.apache.org/)'s properties, through the various Kafka versions, splitting them by categories (broker, topic, producer levels and so forth).

It's all done though a single html page and some static json containing the properties (one per Kakfa version): no backend is required.

Some Go code allows to generate the json file containing the properties per version.

The result is visible on https://kafka-doc.moum.it
